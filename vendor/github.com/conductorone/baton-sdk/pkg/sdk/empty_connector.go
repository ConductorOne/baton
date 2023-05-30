package sdk

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type emptyConnector struct{}

// GetAsset gets an asset.
func (n *emptyConnector) GetAsset(request *v2.AssetServiceGetAssetRequest, server v2.AssetService_GetAssetServer) error {
	err := server.Send(&v2.AssetServiceGetAssetResponse{
		Msg: &v2.AssetServiceGetAssetResponse_Metadata_{
			Metadata: &v2.AssetServiceGetAssetResponse_Metadata{ContentType: "application/example"},
		},
	})
	if err != nil {
		return err
	}

	err = server.Send(&v2.AssetServiceGetAssetResponse{
		Msg: &v2.AssetServiceGetAssetResponse_Data_{
			Data: &v2.AssetServiceGetAssetResponse_Data{Data: nil},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// ListResourceTypes returns a list of resource types.
func (n *emptyConnector) ListResourceTypes(ctx context.Context, request *v2.ResourceTypesServiceListResourceTypesRequest) (*v2.ResourceTypesServiceListResourceTypesResponse, error) {
	return &v2.ResourceTypesServiceListResourceTypesResponse{
		List: []*v2.ResourceType{},
	}, nil
}

// ListResources returns a list of resources.
func (n *emptyConnector) ListResources(ctx context.Context, request *v2.ResourcesServiceListResourcesRequest) (*v2.ResourcesServiceListResourcesResponse, error) {
	return &v2.ResourcesServiceListResourcesResponse{
		List: []*v2.Resource{},
	}, nil
}

// ListEntitlements returns a list of entitlements.
func (n *emptyConnector) ListEntitlements(ctx context.Context, request *v2.EntitlementsServiceListEntitlementsRequest) (*v2.EntitlementsServiceListEntitlementsResponse, error) {
	return &v2.EntitlementsServiceListEntitlementsResponse{
		List: []*v2.Entitlement{},
	}, nil
}

// ListGrants returns a list of grants.
func (n *emptyConnector) ListGrants(ctx context.Context, request *v2.GrantsServiceListGrantsRequest) (*v2.GrantsServiceListGrantsResponse, error) {
	return &v2.GrantsServiceListGrantsResponse{
		List: []*v2.Grant{},
	}, nil
}

// GetMetadata returns a connector metadata.
func (n *emptyConnector) GetMetadata(ctx context.Context, request *v2.ConnectorServiceGetMetadataRequest) (*v2.ConnectorServiceGetMetadataResponse, error) {
	return &v2.ConnectorServiceGetMetadataResponse{Metadata: &v2.ConnectorMetadata{}}, nil
}

// Validate is called by the connector framework to validate the correct response.
func (n *emptyConnector) Validate(ctx context.Context, request *v2.ConnectorServiceValidateRequest) (*v2.ConnectorServiceValidateResponse, error) {
	return &v2.ConnectorServiceValidateResponse{}, nil
}

// NewEmptyConnector returns a new emptyConnector.
func NewEmptyConnector() (*emptyConnector, error) {
	return &emptyConnector{}, nil
}
