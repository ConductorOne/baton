package dotc1z

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type grantJoinKeys struct {
	EntitlementID             string
	EntitlementResourceTypeID string
	EntitlementResourceID     string
	PrincipalResourceTypeID   string
	PrincipalResourceID       string
}

// grantKeys[i] must describe grants[i] — same length, same order.
func hydrateGrants(grants []*v2.Grant, grantKeys []grantJoinKeys) {
	for i, g := range grants {
		entNil := g.GetEntitlement() == nil
		princNil := g.GetPrincipal() == nil
		if !entNil && !princNil {
			continue
		}
		k := grantKeys[i]
		if entNil {
			g.SetEntitlement(stubEntitlement(k))
		}
		if princNil {
			g.SetPrincipal(stubPrincipal(k))
		}
	}
}

func stubEntitlement(k grantJoinKeys) *v2.Entitlement {
	return v2.Entitlement_builder{
		Id: k.EntitlementID,
		Resource: v2.Resource_builder{
			Id: v2.ResourceId_builder{
				ResourceType: k.EntitlementResourceTypeID,
				Resource:     k.EntitlementResourceID,
			}.Build(),
		}.Build(),
	}.Build()
}

func stubPrincipal(k grantJoinKeys) *v2.Resource {
	return v2.Resource_builder{
		Id: v2.ResourceId_builder{
			ResourceType: k.PrincipalResourceTypeID,
			Resource:     k.PrincipalResourceID,
		}.Build(),
	}.Build()
}

// Treats sql.ErrNoRows and empty syncID as soft misses — leaves nil
// fields intact and returns nil. The row may have been deleted between
// GetGrant and this call.
func hydrateSingleGrant(ctx context.Context, c *C1File, syncID string, g *v2.Grant) error {
	if syncID == "" {
		return nil
	}

	row := c.db.QueryRowContext(ctx, fmt.Sprintf(
		`SELECT entitlement_id, resource_type_id, resource_id, principal_resource_type_id, principal_resource_id
		 FROM %s WHERE external_id = ? AND sync_id = ?`, grants.Name(),
	), g.GetId(), syncID)
	var k grantJoinKeys
	if err := row.Scan(
		&k.EntitlementID,
		&k.EntitlementResourceTypeID,
		&k.EntitlementResourceID,
		&k.PrincipalResourceTypeID,
		&k.PrincipalResourceID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("hydrateSingleGrant: fetch identity columns for %q: %w", g.GetId(), err)
	}
	hydrateGrants([]*v2.Grant{g}, []grantJoinKeys{k})
	return nil
}
