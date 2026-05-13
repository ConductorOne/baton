package dotc1z

import "context"

// FileOps is the file-level operations sub-store of C1ZStore. It covers
// operations that span sync runs or produce new sync runs within the same
// c1z file. Cross-file operations (SQL ATTACH) live in pkg/synccompactor,
// not here.
type FileOps interface {
	// CloneSync materializes the given sync run's data into a freshly
	// created standalone c1z file at outPath. Used by c1's storeCompletedSyncC1Z
	// activity to archive a completed sync.
	CloneSync(ctx context.Context, outPath string, syncID string) error

	// GenerateSyncDiff computes the diff between two existing sync runs
	// in this same file and writes the delta as a new SyncTypePartial
	// sync. Returns the new sync's id. Used by the local differ CLI.
	GenerateSyncDiff(ctx context.Context, baseSyncID, appliedSyncID string) (diffSyncID string, err error)
}
