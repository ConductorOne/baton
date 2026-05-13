package dotc1z

import (
	"context"
	"iter"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

// GrantStore is the grant-specific slice of C1ZStore. Each method maps to a
// specific caller intent in the sync pipeline — no mode enums, no option
// structs with behavior switches.
//
// The expansion-reading methods come in two flavors: a page-at-a-time
// primitive (PendingExpansionPage, ListWithAnnotationsPage) that caller-
// driven state machines use to checkpoint progress, and an all-pages
// iterator (PendingExpansion, ListWithAnnotations) that wraps the page
// primitive for callers who want to walk everything in one pass.
type GrantStore interface {
	// StoreExpandedGrants writes grants produced by the expander back to
	// storage. The expander has already consumed each grant's
	// GrantExpandable annotation and performed expansion; storage must
	// persist the grant payload without re-extracting expansion metadata
	// (otherwise we'd corrupt expansion state that has already been
	// marked "done"). For robustness, any residual GrantExpandable
	// annotation on the input is stripped from the persisted payload.
	//
	// Called by pkg/sync/expand.PutGrantsInChunks.
	StoreExpandedGrants(ctx context.Context, grants ...*v2.Grant) error

	// PendingExpansionPage returns the next page of grants whose
	// expansion metadata still needs processing (needs_expansion=1).
	// Grant payloads are NOT materialized — only identity plus the
	// parsed expansion annotation — because the caller only needs
	// expansion metadata and this keeps the hot path cheap.
	//
	// Caller drives pagination by passing the prior nextPageToken as
	// pageToken; an empty nextPageToken means the page was the last.
	// Initial call passes pageToken="".
	//
	// Called by pkg/sync.syncer.ExpandGrants, which checkpoints page
	// tokens in its action state.
	PendingExpansionPage(ctx context.Context, pageToken string) (defs []PendingExpansion, nextPageToken string, err error)

	// PendingExpansion walks all pages of PendingExpansionPage and
	// yields each row. Convenience wrapper for callers that don't
	// checkpoint pagination.
	//
	// ITERATION CONTRACT: on error, the sequence yields a final
	// (zero-value PendingExpansion, err) pair and terminates. Callers
	// MUST check the second return value on every iteration:
	//
	//     for pe, err := range store.Grants().PendingExpansion(ctx) {
	//         if err != nil { return err }
	//         // ... use pe ...
	//     }
	//
	// Single-variable ranging (for pe := range seq) silently drops
	// errors and is a bug.
	PendingExpansion(ctx context.Context) iter.Seq2[PendingExpansion, error]

	// ListWithAnnotationsPage returns the next page of grants with
	// their expansion annotations inline. Grant payloads are fully
	// materialized. Used by the external-principal post-processing
	// step which needs full grants plus expansion.
	//
	// Caller drives pagination by passing the prior nextPageToken as
	// pageToken.
	ListWithAnnotationsPage(ctx context.Context, pageToken string) (rows []GrantAnnotation, nextPageToken string, err error)

	// ListWithAnnotationsForResourcePage filters ListWithAnnotationsPage
	// to grants on the given resource. Used by the c1-side
	// fileClientWrapper that emulates a connector from a c1z file and
	// forwards a ListGrants RPC whose request has a Resource filter.
	//
	// Page size defaults to the engine's maximum when pageSize == 0.
	ListWithAnnotationsForResourcePage(
		ctx context.Context,
		resource *v2.Resource,
		syncID string,
		pageToken string,
		pageSize uint32,
	) (rows []GrantAnnotation, nextPageToken string, err error)

	// ListWithAnnotations walks all pages of ListWithAnnotationsPage
	// and yields each row. Convenience wrapper for callers that don't
	// checkpoint pagination.
	//
	// ITERATION CONTRACT: same as PendingExpansion. On error, the
	// sequence yields (zero-value GrantAnnotation, err) then
	// terminates. Callers MUST check err on every iteration.
	//
	// Called by pkg/sync.syncer.listAllGrantsWithExpansion.
	ListWithAnnotations(ctx context.Context) iter.Seq2[GrantAnnotation, error]
}

// PendingExpansion is a lightweight row yielded by GrantStore.PendingExpansion.
// It carries grant identity and the parsed expansion annotation without
// materializing the full grant payload — this keeps the expansion-worker
// hot path out of proto.Unmarshal.
//
// The syncer uses GrantExternalID to look up the grant and uses Annotation
// to decide what expansion work to enqueue.
type PendingExpansion struct {
	// GrantExternalID is the external id of the grant this expansion
	// applies to (matches v2.Grant.Id).
	GrantExternalID string

	// TargetEntitlementID is the entitlement id granted to the principal.
	TargetEntitlementID string

	// PrincipalResourceTypeID and PrincipalResourceID identify the principal.
	PrincipalResourceTypeID string
	PrincipalResourceID     string

	// Annotation is the grant's expansion annotation. Non-nil by
	// construction: a row only appears in PendingExpansion if it has one.
	Annotation *v2.GrantExpandable

	// NeedsExpansion is the current needs_expansion column value.
	// Always true for rows returned by PendingExpansion — included
	// for parity with the underlying row shape and future flexibility.
	NeedsExpansion bool
}

// GrantAnnotation is a row yielded by GrantStore.ListWithAnnotations.
//
// Grant is always populated. Annotation is nil when the grant has no
// GrantExpandable annotation.
//
// Identity fields (GrantExternalID, TargetEntitlementID,
// PrincipalResourceTypeID, PrincipalResourceID) are ALWAYS populated
// from the underlying grant proto, regardless of whether Annotation is
// nil, so callers can use them without branching.
//
// NeedsExpansion is only meaningful when Annotation is non-nil; it
// reports the stored needs_expansion column for expandable grants.
// For non-expandable grants it is false.
type GrantAnnotation struct {
	Grant      *v2.Grant
	Annotation *v2.GrantExpandable // nil if not expandable

	GrantExternalID         string
	TargetEntitlementID     string
	PrincipalResourceTypeID string
	PrincipalResourceID     string
	NeedsExpansion          bool
}
