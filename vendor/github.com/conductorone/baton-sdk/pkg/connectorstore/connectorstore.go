package connectorstore

import (
	"context"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
)

type SyncType string

const (
	SyncTypeFull             SyncType = "full"
	SyncTypePartial          SyncType = "partial"
	SyncTypeResourcesOnly    SyncType = "resources_only"
	SyncTypePartialUpserts   SyncType = "partial_upserts"   // Diff sync: additions and modifications
	SyncTypePartialDeletions SyncType = "partial_deletions" // Diff sync: deletions
	SyncTypeAny              SyncType = ""
)

var AllSyncTypes = []SyncType{
	SyncTypeAny,
	SyncTypeFull,
	SyncTypePartial,
	SyncTypeResourcesOnly,
	SyncTypePartialUpserts,
	SyncTypePartialDeletions,
}

// ConnectorStoreReader implements the ConnectorV2 API, along with getters for individual objects.
type Reader interface {
	v2.ResourceTypesServiceServer
	reader_v2.ResourceTypesReaderServiceServer

	v2.ResourcesServiceServer
	reader_v2.ResourcesReaderServiceServer

	v2.EntitlementsServiceServer
	reader_v2.EntitlementsReaderServiceServer

	v2.GrantsServiceServer
	reader_v2.GrantsReaderServiceServer

	reader_v2.SyncsReaderServiceServer

	// GetAsset does not implement the AssetServer on the reader here. In other situations we were able to easily 'fake'
	// the GRPC api, but because this is defined as a streaming RPC, it isn't trivial to implement grpc streaming as part of the c1z format.
	GetAsset(ctx context.Context, req *v2.AssetServiceGetAssetRequest) (string, io.Reader, error)

	Close(ctx context.Context) error
}

type InternalWriter interface {
	Writer
	// UpsertGrants writes grants with explicit conflict handling semantics.
	// This is for internal sync workflows that need control over if-newer behavior
	// and whether expansion columns are preserved.
	UpsertGrants(ctx context.Context, opts GrantUpsertOptions, grants ...*v2.Grant) error
	// ListGrantsInternal is the preferred internal listing API for grants.
	// It returns a single list of rows with optional grant payload and expansion metadata.
	ListGrantsInternal(ctx context.Context, opts GrantListOptions) (*InternalGrantListResponse, error)
	// SetSupportsDiff marks the sync as supporting diff operations.
	SetSupportsDiff(ctx context.Context, syncID string) error
}

// GrantUpsertMode controls how grant conflicts are resolved during upsert.
type GrantUpsertMode int

const (
	// GrantUpsertModeReplace updates conflicting grants unconditionally.
	GrantUpsertModeReplace GrantUpsertMode = iota
	// GrantUpsertModeIfNewer updates conflicting grants only when EXCLUDED.discovered_at is newer.
	GrantUpsertModeIfNewer
	// GrantUpsertModePreserveExpansion updates grant data while preserving existing
	// expansion and needs_expansion columns.
	GrantUpsertModePreserveExpansion
)

// GrantUpsertOptions configures internal grant upsert behavior.
type GrantUpsertOptions struct {
	Mode GrantUpsertMode
}

// ConnectorStoreWriter defines an implementation for a connector v2 datasource writer. This is used to store sync data from an upstream provider.
type Writer interface {
	Reader
	ResumeSync(ctx context.Context, syncType SyncType, syncID string) (string, error)
	StartOrResumeSync(ctx context.Context, syncType SyncType, syncID string) (string, bool, error)
	StartNewSync(ctx context.Context, syncType SyncType, parentSyncID string) (string, error)
	SetCurrentSync(ctx context.Context, syncID string) error
	CurrentSyncStep(ctx context.Context) (string, error)
	CheckpointSync(ctx context.Context, syncToken string) error
	EndSync(ctx context.Context) error
	PutAsset(ctx context.Context, assetRef *v2.AssetRef, contentType string, data []byte) error
	Cleanup(ctx context.Context) error

	PutGrants(ctx context.Context, grants ...*v2.Grant) error
	PutResourceTypes(ctx context.Context, resourceTypes ...*v2.ResourceType) error
	PutResources(ctx context.Context, resources ...*v2.Resource) error
	PutEntitlements(ctx context.Context, entitlements ...*v2.Entitlement) error
	DeleteGrant(ctx context.Context, grantId string) error
}

// GrantListMode configures which row shape/filter mode ListGrantsInternal uses.
type GrantListMode int

const (
	// GrantListModePayload returns grant payload rows only.
	GrantListModePayload GrantListMode = iota
	// GrantListModePayloadWithExpansion returns grant payload rows with optional expansion metadata.
	GrantListModePayloadWithExpansion
	// GrantListModeExpansion returns expansion metadata rows only.
	GrantListModeExpansion
	// GrantListModeExpansionNeedsOnly returns only expansion metadata rows with needs_expansion=1.
	GrantListModeExpansionNeedsOnly
)

// GrantListOptions configures ListGrantsInternal - a would be union type.
type GrantListOptions struct {
	// Mode controls which row shape/filter is returned.
	Mode GrantListMode

	// Resource filters payload modes to grants on a specific resource.
	Resource *v2.Resource

	// ExpandableOnly filters rows to grants with expansion metadata.
	// Used by payload+expansion mode.
	ExpandableOnly bool
	// NeedsExpansionOnly filters rows to needs_expansion=1.
	// Used by expansion-only modes.
	NeedsExpansionOnly bool

	// PageToken and PageSize are used for pagination in all modes.
	// SyncID is used for expansion-only modes.
	SyncID    string
	PageToken string
	PageSize  uint32
}

// InternalGrantRow is one row from ListGrantsInternal. Fields are optional
// based on the requested list options.
type InternalGrantRow struct {
	Grant     *v2.Grant
	Expansion *ExpandableGrantDef
}

// InternalGrantListResponse contains one row list plus a shared next page token.
type InternalGrantListResponse struct {
	Rows          []*InternalGrantRow
	NextPageToken string
}

// ExpandableGrantDef is a lightweight representation of an expandable grant row,
// using queryable columns instead of unmarshalling the full grant proto.
type ExpandableGrantDef struct {
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
