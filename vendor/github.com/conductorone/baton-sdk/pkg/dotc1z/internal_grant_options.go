package dotc1z

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

// These types are unexported implementation details of *C1File's grant-store
// behavior. They replace the former connectorstore.GrantUpsertOptions,
// connectorstore.GrantListOptions, and related types that used to be part
// of the connector-facing surface. Nothing outside this package should
// reference them; RFC 0002 removed the InternalWriter interface so callers
// now go through GrantStore (grant_store.go) instead.

// grantUpsertMode controls how grant conflicts are resolved during upsert.
type grantUpsertMode int

const (
	// grantUpsertModeReplace updates conflicting grants unconditionally.
	grantUpsertModeReplace grantUpsertMode = iota
	// grantUpsertModeIfNewer updates conflicting grants only when EXCLUDED.discovered_at is newer.
	grantUpsertModeIfNewer
	// grantUpsertModePreserveExpansion updates grant data while preserving existing
	// expansion and needs_expansion columns.
	grantUpsertModePreserveExpansion
)

// grantUpsertOptions configures internal grant upsert behavior.
type grantUpsertOptions struct {
	Mode grantUpsertMode
}

// grantListMode selects which row shape/filter listGrantsInternal uses.
type grantListMode int

const (
	// grantListModePayloadWithExpansion returns grant payload rows with
	// optional expansion metadata.
	grantListModePayloadWithExpansion grantListMode = iota
	// grantListModeExpansionNeedsOnly returns only expansion metadata rows
	// with needs_expansion=1.
	grantListModeExpansionNeedsOnly
)

// grantListOptions configures listGrantsInternal.
type grantListOptions struct {
	// Mode selects the row shape/filter.
	Mode grantListMode

	// Resource filters payload modes to grants on a specific resource.
	Resource *v2.Resource

	// ExpandableOnly filters rows to grants with expansion metadata.
	ExpandableOnly bool

	// SyncID is used for expansion-only modes.
	SyncID    string
	PageToken string
	PageSize  uint32
}

// internalGrantRow is one row from listGrantsInternal. Fields are optional
// based on the requested list options.
type internalGrantRow struct {
	Grant     *v2.Grant
	Expansion *expandableGrantDef
}

// internalGrantListResponse contains one row list plus a shared next page token.
type internalGrantListResponse struct {
	Rows          []*internalGrantRow
	NextPageToken string
}

// expandableGrantDef is a lightweight representation of an expandable grant
// row, using queryable columns instead of unmarshalling the full grant proto.
// This is the raw storage-layer shape; callers that want the exported form
// use GrantStore.PendingExpansionPage which yields PendingExpansion.
type expandableGrantDef struct {
	RowID                   int64
	GrantExternalID         string
	TargetEntitlementID     string
	PrincipalResourceTypeID string
	PrincipalResourceID     string
	SourceEntitlementIDs    []string
	Shallow                 bool
	ResourceTypeIDs         []string
	NeedsExpansion          bool
}
