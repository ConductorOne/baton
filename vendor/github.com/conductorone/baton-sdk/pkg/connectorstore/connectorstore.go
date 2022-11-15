package connectorstore

import (
	"context"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
)

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

	// GetAsset does not implement the AssetServer on the reader here. In other situations we were able to easily 'fake'
	// the GRPC api, but because this is defined as a streaming RPC, it isn't trivial to implement grpc streaming as part of the c1z format.
	GetAsset(ctx context.Context, req *v2.AssetServiceGetAssetRequest) (string, io.Reader, error)

	// ViewSync uses the provided syncID to change which sync generation is used for fetching results.
	// If this is not called, the latest complete sync will be used.
	ViewSync(ctx context.Context, syncID string) error

	Close() error
}

// ConnectorStoreWriter defines an implementation for a connector v2 datasource writer. This is used to store sync data from an upstream provider.
type Writer interface {
	Reader
	StartSync(ctx context.Context) (string, bool, error)
	CurrentSyncStep(ctx context.Context) (string, error)
	CheckpointSync(ctx context.Context, syncToken string) error
	EndSync(ctx context.Context) error
	PutResourceType(ctx context.Context, resourceType *v2.ResourceType) error
	PutResource(ctx context.Context, resource *v2.Resource) error
	PutEntitlement(ctx context.Context, entitlement *v2.Entitlement) error
	PutGrant(ctx context.Context, grant *v2.Grant) error
	PutAsset(ctx context.Context, assetRef *v2.AssetRef, contentType string, data []byte) error
}
