package dotc1z

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/connectorstore"
)

// C1ZStore is the internal contract used by the sync pipeline, compactor, and
// related infrastructure to read and write a .c1z file. It is NOT the
// connector-facing contract — connectors write through connectorstore.Writer,
// which is embedded here but stays the stable API surface for external code.
//
// C1ZStore adds three kinds of operations that the sync pipeline needs but
// connectors don't:
//
//   - Grant operations with expansion-aware semantics, accessed via Grants().
//   - Sync-run metadata operations, accessed via SyncMeta().
//   - File-level operations (clone, diff), accessed via FileOps().
//
// *C1File is the sole implementation.
//
// The Close override replaces connectorstore.Reader.Close with an
// explicit context parameter (*C1File.Close already takes context on main,
// so this is exact match, no behavior change).
type C1ZStore interface {
	connectorstore.Writer

	// Grants returns the grant-specific sub-store.
	Grants() GrantStore

	// SyncMeta returns the sync-run metadata sub-store.
	SyncMeta() SyncMeta

	// FileOps returns the file-level operations sub-store.
	FileOps() FileOps

	// Close releases resources, flushing any pending writes.
	// Overrides the embedded Reader.Close to document this signature.
	Close(ctx context.Context) error
}

// AsSQLiteStore type-asserts a C1ZStore to the concrete *C1File. It is an
// escape hatch for callers that legitimately need SQLite-specific primitives
// (today: the attached compactor in pkg/synccompactor/attached, which uses
// SQL ATTACH for cross-file merge). Returns (nil, false) when the store is
// not backed by *C1File OR when the underlying *C1File is nil.
//
// Avoid using this outside pkg/synccompactor. If you find yourself reaching
// for it, prefer adding a named method to C1ZStore that expresses what you
// need; sqlite-specific leak-through is a smell. See RFC 0002 §4.4.
func AsSQLiteStore(s C1ZStore) (*C1File, bool) {
	cf, ok := s.(*C1File)
	if !ok || cf == nil {
		return nil, false
	}
	return cf, true
}
