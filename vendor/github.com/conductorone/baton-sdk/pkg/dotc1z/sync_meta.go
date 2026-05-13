package dotc1z

import (
	"context"
	"time"

	"github.com/conductorone/baton-sdk/pkg/connectorstore"
)

// SyncMeta is the sync-run-metadata sub-store of C1ZStore. It covers
// operations that read/write the sync_runs table without going through the
// gRPC Reader/Writer surfaces.
//
// All methods are callable without an active sync.
type SyncMeta interface {
	// MarkSyncSupportsDiff sets the supports_diff flag on the given sync.
	// Called by pkg/sync.parallelSyncer after graph construction to signal
	// that the sync run has SQL-layer grant metadata populated and diff
	// consumers may rely on it.
	MarkSyncSupportsDiff(ctx context.Context, syncID string) error

	// LatestFullSync returns the most-recently-finished SyncTypeFull sync
	// run, or nil if none exists. Used by pkg/sync.syncer to compute the
	// "previous sync" reference for etag-based grant reuse.
	LatestFullSync(ctx context.Context) (*SyncRun, error)

	// LatestFinishedSyncOfAnyType returns the most-recently-finished sync
	// of any type (including diff types), or nil if none exists. Used by
	// tooling that wants to inspect whatever sync finished last regardless
	// of type.
	LatestFinishedSyncOfAnyType(ctx context.Context) (*SyncRun, error)

	// Stats returns a map of table-name to row-count for the given sync.
	// If syncID is empty, the latest sync of the given type is used.
	// Mirrors the existing *C1File.Stats signature exactly.
	Stats(ctx context.Context, syncType connectorstore.SyncType, syncID string) (map[string]int64, error)
}

// SyncRun is the exported shape of a sync run. It corresponds directly to the
// internal syncRun struct; the fields match the sync_runs schema.
//
// Callers typically only read ID, Type, and the timestamps; the rest is
// included for completeness and for use by tooling (e.g. sync-diff
// pipelines need ParentSyncID and LinkedSyncID).
type SyncRun struct {
	ID           string
	StartedAt    *time.Time
	EndedAt      *time.Time
	SyncToken    string
	Type         connectorstore.SyncType
	ParentSyncID string
	LinkedSyncID string
	SupportsDiff bool
}
