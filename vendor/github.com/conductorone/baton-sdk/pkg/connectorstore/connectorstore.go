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

// LatestFinishedSyncIDFetcher returns the most-recently-finished sync ID of the
// given type, or empty string if no such sync exists. This is a small optional
// capability separate from Reader/Writer because not every store implementation
// can answer it (e.g. gRPC-backed readers have a different flavor via
// SyncsReaderServiceGetLatestFinishedSync).
//
// This interface lives in connectorstore so that producers (e.g. *dotc1z.C1File)
// and consumers (e.g. pkg/sync) reference a single authoritative declaration,
// preventing the name/signature drift that occurred between PR #473 and RFC 0002.
type LatestFinishedSyncIDFetcher interface {
	LatestFinishedSyncID(ctx context.Context, syncType SyncType) (string, error)
}

// DBSizeProvider is an optional capability for a store that can report its
// current uncompressed working-set size (e.g. dotc1z.C1File stat'ing its
// sqlite file). Consumed by the syncer's ProgressLog to include
// decompressed_bytes and growth delta in the periodic "Expanding grants"
// log during long-running grant expansions — the Expander itself is
// recreated each RunSingleStep by the syncer, so this state cannot live
// there.
type DBSizeProvider interface {
	CurrentDBSizeBytes() (int64, error)
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
