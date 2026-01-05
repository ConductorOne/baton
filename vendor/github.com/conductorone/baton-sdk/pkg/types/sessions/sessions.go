package sessions

import (
	"context"
)

const MaxKeysPerRequest = 100

// The default gRPC message size limit is 4MB (we subtract 30KB for general overhead, which is overkill).
// Unfortunately, this layer has to be aware of the size limit to avoid exceeding the size limit
// because the client does not know the size of the items it requests.
const MaxSessionStoreSizeLimit = 4163584

type SessionStoreKey struct{}

type SessionStore interface {
	Get(ctx context.Context, key string, opt ...SessionStoreOption) ([]byte, bool, error)
	GetMany(ctx context.Context, keys []string, opt ...SessionStoreOption) (map[string][]byte, []string, error)
	Set(ctx context.Context, key string, value []byte, opt ...SessionStoreOption) error
	SetMany(ctx context.Context, values map[string][]byte, opt ...SessionStoreOption) error
	Delete(ctx context.Context, key string, opt ...SessionStoreOption) error
	Clear(ctx context.Context, opt ...SessionStoreOption) error
	GetAll(ctx context.Context, pageToken string, opt ...SessionStoreOption) (map[string][]byte, string, error)
}

type SessionStoreOption func(ctx context.Context, bag *SessionStoreBag) error

type SessionStoreConstructor func(ctx context.Context, opt ...SessionStoreConstructorOption) (SessionStore, error)

type SessionStoreConstructorOption func(ctx context.Context) (context.Context, error)

type SessionStoreBag struct {
	SyncID    string
	Prefix    string
	PageToken string
}

// SyncIDKey is the context key for storing the current sync ID.
type SyncIDKey struct{}

// WithSyncID returns a SessionCacheOption that sets the sync ID for the operation.
func WithSyncID(syncID string) SessionStoreOption {
	return func(ctx context.Context, bag *SessionStoreBag) error {
		bag.SyncID = syncID
		return nil
	}
}

func WithPrefix(prefix string) SessionStoreOption {
	return func(ctx context.Context, bag *SessionStoreBag) error {
		bag.Prefix = prefix
		return nil
	}
}

// GetSyncID retrieves the sync ID from the context, returning empty string if not found.
func GetSyncID(ctx context.Context) string {
	if syncID, ok := ctx.Value(SyncIDKey{}).(string); ok {
		return syncID
	}
	return ""
}

func WithPageToken(pageToken string) SessionStoreOption {
	return func(ctx context.Context, bag *SessionStoreBag) error {
		bag.PageToken = pageToken
		return nil
	}
}

func SetSyncIDInContext(ctx context.Context, syncID string) context.Context {
	return context.WithValue(ctx, SyncIDKey{}, syncID)
}

type SetSessionStore interface {
	SetSessionStore(ctx context.Context, store SessionStore)
}
