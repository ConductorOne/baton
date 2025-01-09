package uhttp

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

	"github.com/doug-martin/goqu/v9"
	// NOTE: required to register the dialect for goqu.
	//
	// If you remove this import, goqu.Dialect("sqlite3") will
	// return a copy of the default dialect, which is not what we want,
	// and allocates a ton of memory.
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/glebarez/go-sqlite"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type DBCache struct {
	rawDb *sql.DB
	db    *goqu.Database
	// Cleanup interval, close and remove db
	waitDuration time.Duration
	// Cache duration for removing expired keys
	expirationTime time.Duration
	// Database path
	location string
	// Enable statistics(hits, misses)
	stats bool
}

type CacheRow struct {
	Key        string
	Value      []byte
	Expires    time.Time
	LastAccess time.Time
	Url        string
}

const (
	failStartTransaction        = "Failed to start a transaction"
	errQueryingTable            = "Error querying cache table"
	failRollback                = "Failed to rollback transaction"
	failInsert                  = "Failed to insert response data into cache table"
	failScanResponse            = "Failed to scan rows for cached response"
	cacheTTLThreshold           = 60
	cacheTTLMultiplier   uint64 = 5
)

var errNilConnection = errors.New("database connection is nil")

var defaultWaitDuration = cacheTTLThreshold * time.Second // Default Cleanup interval, 60 seconds

const tableName = "http_cache"

// TODO (ggreer): obey c1z-temp-dir CLI arg or environment variable
func NewDBCache(ctx context.Context, cfg CacheConfig) (*DBCache, error) {
	var (
		err error
		dc  = &DBCache{
			waitDuration: defaultWaitDuration, // Default Cleanup interval, 60 seconds
			stats:        true,
			//nolint:gosec // disable G115
			expirationTime: time.Duration(cfg.TTL) * time.Second,
		}
	)
	l := ctxzap.Extract(ctx)
	dc, err = dc.load(ctx)
	if err != nil {
		l.Warn("Failed to open database", zap.Error(err))
		return nil, err
	}

	// Create cache table and index
	_, err = dc.db.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS http_cache(
		key TEXT PRIMARY KEY, 
		value BLOB, 
		expires INT, 
		lastAccess TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		url TEXT
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_cache_key ON http_cache (key);
	CREATE INDEX IF NOT EXISTS expires ON http_cache (expires);
	CREATE INDEX IF NOT EXISTS lastAccess ON http_cache (lastAccess);
	CREATE TABLE IF NOT EXISTS http_stats(
		id INTEGER PRIMARY KEY,
		key TEXT,
		hits INT DEFAULT 0, 
		misses INT DEFAULT 0
	);
	DELETE FROM http_cache;
	DELETE FROM http_stats;`)
	if err != nil {
		l.Warn("Failed to create cache table in database", zap.Error(err))
		return nil, err
	}

	if cfg.TTL <= 0 {
		l.Debug("Cache TTL is 0. Disabling cache.")
		return nil, nil
	}

	if cfg.TTL > cacheTTLThreshold {
		//nolint:gosec // disable G115
		dc.waitDuration = time.Duration(cfg.TTL*cacheTTLMultiplier) * time.Second // set as a fraction of the Cache TTL
	}

	go func(waitDuration, expirationTime time.Duration) {
		ctxWithTimeout, cancel := context.WithTimeout(
			ctx,
			waitDuration,
		)
		defer cancel()
		// TODO: I think this should be wait duration
		ticker := time.NewTicker(expirationTime)
		defer ticker.Stop()
		for {
			select {
			case <-ctxWithTimeout.Done():
				// ctx done, shutting down cache cleanup routine
				ticker.Stop()
				err := dc.cleanup(ctx)
				if err != nil {
					l.Warn("shutting down cache failed", zap.Error(err))
				}
				return
			case <-ticker.C:
				err := dc.deleteExpired(ctx)
				if err != nil {
					l.Warn("Failed to delete expired cache entries", zap.Error(err))
				}
			}
		}
	}(dc.waitDuration, dc.expirationTime)

	return dc, nil
}

func (d *DBCache) load(ctx context.Context) (*DBCache, error) {
	l := ctxzap.Extract(ctx)
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		l.Warn("Failed to read user cache directory", zap.Error(err))
		return nil, err
	}

	file := filepath.Join(cacheDir, "lcache.db")
	d.location = file

	l.Debug("Opening database", zap.String("location", file))
	rawDB, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, err
	}
	l.Debug("Opened database", zap.String("location", file))

	d.db = goqu.New("sqlite3", rawDB)
	d.rawDb = rawDB
	return d, nil
}

func (d *DBCache) removeDB(ctx context.Context) error {
	_, err := os.Stat(d.location)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file not found %s", d.location)
	}

	err = d.close(ctx)
	if err != nil {
		return err
	}

	err = os.Remove(d.location)
	if err != nil {
		ctxzap.Extract(ctx).Warn("error removing cache database", zap.Error(err))
		return err
	}

	return nil
}

// Get returns cached response (if exists).
func (d *DBCache) Get(req *http.Request) (*http.Response, error) {
	var (
		isFound bool = false
		resp    *http.Response
	)
	key, err := CreateCacheKey(req)
	if err != nil {
		return nil, err
	}
	ctx := req.Context()
	l := ctxzap.Extract(ctx)

	entry, err := d.pick(ctx, key)
	if err == nil && len(entry) > 0 {
		r := bufio.NewReader(bytes.NewReader(entry))
		resp, err = http.ReadResponse(r, nil)
		if err != nil {
			return nil, err
		}

		isFound = true
	}

	field := "misses"
	if isFound {
		field = "hits"
	}
	err = d.updateStats(ctx, field, key)
	if err != nil {
		l.Warn("Failed to update cache stats", zap.Error(err), zap.String("field", field))
	}

	return resp, nil
}

func (d *DBCache) pick(ctx context.Context, key string) ([]byte, error) {
	if d.db == nil {
		return nil, errNilConnection
	}

	l := ctxzap.Extract(ctx)
	ds := goqu.From(tableName).Select("value").Where(goqu.Ex{"key": key})
	query, args, err := ds.ToSQL()
	if err != nil {
		l.Warn("Failed to create select statement", zap.Error(err))
		return nil, err
	}

	var value []byte
	row := d.db.QueryRowContext(ctx, query, args...)
	err = row.Scan(&value)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		l.Warn(errQueryingTable, zap.Error(err), zap.String("sql", query), zap.Any("args", args))
		return nil, err
	}

	return value, nil
}

// Set stores and save response in the db.
func (d *DBCache) Set(req *http.Request, value *http.Response) error {
	key, err := CreateCacheKey(req)
	if err != nil {
		return err
	}
	cacheableResponse, err := httputil.DumpResponse(value, true)
	if err != nil {
		return err
	}

	url := req.URL.String()

	err = d.insert(req.Context(),
		key,
		cacheableResponse,
		url,
	)
	if err != nil {
		return err
	}

	return nil
}

func (d *DBCache) cleanup(ctx context.Context) error {
	l := ctxzap.Extract(ctx)
	stats, err := d.getStats(ctx)
	if err != nil {
		l.Warn("error getting stats", zap.Error(err))
		return err
	}

	l.Debug("summary and stats", zap.Any("stats", stats))
	return d.removeDB(ctx)
}

// Insert data into the cache table.
func (d *DBCache) insert(ctx context.Context, key string, bytes []byte, url string) error {
	if d.db == nil {
		return errNilConnection
	}

	l := ctxzap.Extract(ctx)

	tx, err := d.db.Begin()
	if err != nil {
		l.Warn(failStartTransaction, zap.Error(err))
		return err
	}

	ds := goqu.Insert(tableName).Rows(
		CacheRow{
			Key:     key,
			Value:   bytes,
			Expires: time.Now().Add(d.expirationTime),
			Url:     url,
		},
	)
	ds = ds.OnConflict(goqu.DoUpdate("key", CacheRow{
		Key:     key,
		Value:   bytes,
		Expires: time.Now().Add(d.expirationTime),
		Url:     url,
	}))
	insertSQL, args, err := ds.ToSQL()
	if err != nil {
		l.Warn("Failed to create insert statement", zap.Error(err))
		return err
	}
	_, err = tx.ExecContext(ctx, insertSQL, args...)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			l.Warn(failRollback, zap.Error(errtx))
		}

		l.Warn(failInsert, zap.Error(err), zap.String("sql", insertSQL))
		return err
	}

	err = tx.Commit()
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			l.Warn(failRollback, zap.Error(errtx))
		}

		l.Warn(failInsert, zap.Error(err), zap.String("sql", insertSQL))
		return err
	}

	return nil
}

// Delete all expired items from the cache.
func (d *DBCache) deleteExpired(ctx context.Context) error {
	if d.db == nil {
		return errNilConnection
	}

	l := ctxzap.Extract(ctx)
	tx, err := d.db.Begin()
	if err != nil {
		l.Warn(failStartTransaction, zap.Error(err))
		return err
	}

	_, err = d.db.ExecContext(ctx, "DELETE FROM http_cache WHERE expires < ?", time.Now().UnixNano())
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			l.Warn(failRollback, zap.Error(errtx))
		}

		l.Warn("Failed to delete expired cache entries", zap.Error(err))
		return err
	}

	err = tx.Commit()
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			l.Warn(failRollback, zap.Error(errtx))
		}
		l.Warn("Failed to commit after deleting expired cache entries", zap.Error(err))
	}

	return err
}

func (d *DBCache) close(ctx context.Context) error {
	if d.rawDb == nil {
		return errNilConnection
	}

	err := d.rawDb.Close()
	if err != nil {
		ctxzap.Extract(ctx).Warn("Failed to close database connection", zap.Error(err))
	}

	return err
}

func (d *DBCache) updateStats(ctx context.Context, field, key string) error {
	if !d.stats {
		return nil
	}

	if d.db == nil {
		return errNilConnection
	}
	l := ctxzap.Extract(ctx)
	tx, err := d.db.Begin()
	if err != nil {
		l.Warn(failStartTransaction, zap.Error(err))
		return err
	}

	_, err = d.db.ExecContext(ctx, fmt.Sprintf("INSERT INTO http_stats(key, %s) values(?, 1)", field), key)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			l.Warn(failRollback, zap.Error(errtx))
		}

		l.Warn("error updating "+field, zap.Error(err))
		return err
	}

	err = tx.Commit()
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			l.Warn(failRollback, zap.Error(errtx))
		}

		l.Warn("Failed to update "+field, zap.Error(err))
		return err
	}

	return nil
}

func (d *DBCache) getStats(ctx context.Context) (CacheStats, error) {
	var (
		hits   int64
		misses int64
	)
	if d.db == nil {
		return CacheStats{}, errNilConnection
	}

	l := ctxzap.Extract(ctx)
	rows, err := d.db.QueryContext(ctx, `
	SELECT 
		sum(hits) total_hits, 
		sum(misses) total_misses 
	FROM http_stats
	`)
	if err != nil {
		l.Warn(errQueryingTable, zap.Error(err))
		return CacheStats{}, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&hits, &misses)
		if err != nil {
			l.Warn(failScanResponse, zap.Error(err))
			return CacheStats{}, err
		}
	}

	return CacheStats{
		Hits:   hits,
		Misses: misses,
	}, nil
}

func (d *DBCache) Stats(ctx context.Context) CacheStats {
	stats, err := d.getStats(ctx)
	if err != nil {
		return CacheStats{}
	}
	return stats
}

func (d *DBCache) Clear(ctx context.Context) error {
	// TODO: Implement
	return nil
}
