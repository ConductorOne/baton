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
