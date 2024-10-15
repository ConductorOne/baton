package dotc1z

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

const assetsTableVersion = "1"
const assetsTableName = "assets"
const assetsTableSchema = `
create table if not exists %s (
    id integer primary key,
    external_id text not null,
    content_type text not null,
    data blob not null,
    sync_id text not null,
    discovered_at datetime not null
);
create unique index if not exists %s on %s (external_id, sync_id);`

var assets = (*assetsTable)(nil)

type assetsTable struct{}

// Name returns the formatted table name for the assets table.
func (r *assetsTable) Name() string {
	return fmt.Sprintf("v%s_%s", r.Version(), assetsTableName)
}

// Version returns the hard coded version of the table.
func (r *assetsTable) Version() string {
	return assetsTableVersion
}

// Schema returns the expanded SQL to generate the assets table.
func (r *assetsTable) Schema() (string, []interface{}) {
	return assetsTableSchema, []interface{}{
		r.Name(),
		fmt.Sprintf("idx_assets_external_sync_v%s", r.Version()),
		r.Name(),
	}
}

// PutAsset stores the given asset in the database.
func (c *C1File) PutAsset(ctx context.Context, assetRef *v2.AssetRef, contentType string, data []byte) error {
	l := ctxzap.Extract(ctx)

	if len(data) == 0 {
		l.Debug("skipping storing empty asset")
		return nil
	}

	if contentType == "" {
		l.Debug("empty content type")
		contentType = "unknown"
	}

	err := c.validateSyncDb(ctx)
	if err != nil {
		return err
	}

	fields := goqu.Record{
		"external_id":   assetRef.Id,
		"content_type":  contentType,
		"data":          data,
		"sync_id":       c.currentSyncID,
		"discovered_at": time.Now().Format("2006-01-02 15:04:05.999999999"),
	}

	q := c.db.Insert(assets.Name()).Prepared(true)
	q = q.Rows(fields)
	q = q.OnConflict(goqu.DoUpdate("external_id, sync_id", goqu.C("data").Set(goqu.I("EXCLUDED.data"))))

	query, args, err := q.ToSQL()
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	c.dbUpdated = true

	return nil
}

// GetAsset fetches the specified asset from the database, and returns the content type and an io.Reader for the caller to
// read the asset from.
func (c *C1File) GetAsset(ctx context.Context, request *v2.AssetServiceGetAssetRequest) (string, io.Reader, error) {
	err := c.validateDb(ctx)
	if err != nil {
		return "", nil, err
	}

	if request.Asset == nil {
		return "", nil, fmt.Errorf("asset is required")
	}

	q := c.db.From(assets.Name()).Prepared(true)
	q = q.Select("content_type", "data")
	q = q.Where(goqu.C("external_id").Eq(request.Asset.Id))

	if c.currentSyncID != "" {
		q = q.Where(goqu.C("sync_id").Eq(c.currentSyncID))
	}

	query, args, err := q.ToSQL()
	if err != nil {
		return "", nil, err
	}

	var contentType string
	data := make([]byte, 0)

	row := c.db.QueryRowContext(ctx, query, args...)
	err = row.Scan(&contentType, &data)
	if err != nil {
		return "", nil, err
	}

	out := bytes.NewBuffer(data)

	return contentType, out, nil
}
