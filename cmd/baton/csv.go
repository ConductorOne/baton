package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorstore"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/cobra"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

func exportCSV() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "csv",
		Short: "Export CSV for upload",
		RunE:  runExportCSV,
	}

	cmd.Flags().String("out", "./sync.csv", "The path to export the CSV to")

	return cmd
}

func fetchResourceTypes(ctx context.Context, store connectorstore.Reader) (map[string]*v2.ResourceType, error) {
	ret := make(map[string]*v2.ResourceType)
	pageToken := ""
	for {
		req := &v2.ResourceTypesServiceListResourceTypesRequest{PageToken: pageToken}
		resp, err := store.ListResourceTypes(ctx, req)
		if err != nil {
			return nil, err
		}

		for _, rt := range resp.List {
			ret[rt.Id] = rt
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return ret, nil
}

// fetchResources returns the resources in two maps: by ID -> resource and resource type ->  ID -> resource.
func fetchResources(ctx context.Context, store connectorstore.Reader) (map[string]*v2.Resource, map[string]map[string]*v2.Resource, error) {
	ret := make(map[string]*v2.Resource)
	retRt := make(map[string]map[string]*v2.Resource)

	pageToken := ""
	for {
		req := &v2.ResourcesServiceListResourcesRequest{PageToken: pageToken}
		resp, err := store.ListResources(ctx, req)
		if err != nil {
			return nil, nil, err
		}

		for _, r := range resp.List {
			m, ok := retRt[r.Id.ResourceType]
			if !ok {
				m = make(map[string]*v2.Resource)
			}

			m[fmtResourceID(r.Id)] = r
			retRt[r.Id.ResourceType] = m
			ret[fmtResourceID(r.Id)] = r
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return ret, retRt, nil
}

func fetchEntitlements(ctx context.Context, store connectorstore.Reader) (map[string]*v2.Entitlement, map[string]map[string]*v2.Entitlement, error) {
	ret := make(map[string]*v2.Entitlement)
	retRt := make(map[string]map[string]*v2.Entitlement)

	pageToken := ""
	for {
		req := &v2.EntitlementsServiceListEntitlementsRequest{PageToken: pageToken}
		resp, err := store.ListEntitlements(ctx, req)
		if err != nil {
			return nil, nil, err
		}

		for _, e := range resp.List {
			m, ok := retRt[fmtResourceID(e.Resource.Id)]
			if !ok {
				m = make(map[string]*v2.Entitlement)
			}

			m[e.Id] = e
			retRt[fmtResourceID(e.Resource.Id)] = m
			ret[e.Id] = e
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return ret, retRt, nil
}

func fetchGrants(ctx context.Context, store connectorstore.Reader) (map[string]*v2.Grant, map[string]map[string]*v2.Grant, error) {
	ret := make(map[string]*v2.Grant)
	retRt := make(map[string]map[string]*v2.Grant)

	pageToken := ""
	for {
		req := &v2.GrantsServiceListGrantsRequest{PageToken: pageToken}
		resp, err := store.ListGrants(ctx, req)
		if err != nil {
			return nil, nil, err
		}

		for _, g := range resp.List {
			m, ok := retRt[fmtResourceID(g.Principal.Id)]
			if !ok {
				m = make(map[string]*v2.Grant)
			}

			m[g.Id] = g
			retRt[fmtResourceID(g.Principal.Id)] = m
			ret[g.Id] = g
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return ret, retRt, nil
}

func runExportCSV(cmd *cobra.Command, args []string) error {
	ctx, err := logging.Init(context.Background(), logging.WithLogFormat("console"), logging.WithLogLevel("error"))
	if err != nil {
		return err
	}
	c1zPath, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	outPath, err := cmd.Flags().GetString("out")
	if err != nil {
		return err
	}

	m, err := manager.New(ctx, c1zPath)
	if err != nil {
		return err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return err
	}

	resourceTypes, err := fetchResourceTypes(ctx, store)
	if err != nil {
		return err
	}
	resourcesByID, resourcesByType, err := fetchResources(ctx, store)
	if err != nil {
		return err
	}
	entitlementsByID, entitlementsByResource, err := fetchEntitlements(ctx, store)
	if err != nil {
		return err
	}
	grantsByID, grantsByPrincipal, err := fetchGrants(ctx, store)
	if err != nil {
		return err
	}

	d := dataBag{
		resourceTypes:      resourceTypes,
		resourcesByID:      resourcesByID,
		resourcesByType:    resourcesByType,
		entitlementsByID:   entitlementsByID,
		entitlementsByType: entitlementsByResource,
		grantsByID:         grantsByID,
		grantsByType:       grantsByPrincipal,
	}

	err = buildCSV(ctx, d, outPath)
	if err != nil {
		return err
	}

	return nil
}

type csvRow struct {
	rowType                string
	lastName               string
	firstName              string
	userID                 string
	userStatus             string
	emailAddress           string
	entitlementDisplayName string
	entitlement            string
	resourceType           string
	resourceName           string
	entitlementDescription string
	entitlementSlug        string
}

func (c csvRow) Row() []string {
	return []string{
		c.rowType,
		c.lastName,
		c.firstName,
		c.userID,
		c.userStatus,
		c.emailAddress,
		c.entitlementDisplayName,
		c.entitlement,
		c.resourceType,
		c.resourceName,
		c.entitlementDescription,
		c.entitlementSlug,
	}
}

func getUserStatus(ctx context.Context, ut *v2.UserTrait) string {
	if ut.Status == nil {
		return "Enabled"
	}

	switch ut.Status.Status {
	case v2.UserTrait_Status_STATUS_ENABLED:
		return "Enabled"
	case v2.UserTrait_Status_STATUS_DISABLED:
		return "Disabled"
	case v2.UserTrait_Status_STATUS_DELETED:
		return "Deleted"
	default:
		return "Unknown"
	}
}

func buildCSV(ctx context.Context, d dataBag, outPath string) error {
	l := ctxzap.Extract(ctx)
	l.Debug("building CSV")

	var err error

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	err = w.Write(headers())
	if err != nil {
		return err
	}

	// Identity
	var userTypes []string
	for _, rt := range d.resourceTypes {
		for _, t := range rt.Traits {
			if t == v2.ResourceType_TRAIT_USER {
				userTypes = append(userTypes, rt.Id)
			}
		}
	}

	for _, userType := range userTypes {
		var ok bool
		var rs map[string]*v2.Resource

		rs, ok = d.resourcesByType[userType]
		if !ok {
			return err
		}

		for _, r := range rs {
			ut := &v2.UserTrait{}
			annos := annotations.Annotations(r.Annotations)
			ok, err = annos.Pick(ut)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}

			var emailAddress string
			for _, e := range ut.Emails {
				if e.IsPrimary {
					emailAddress = e.Address
					break
				}
			}

			profile := ut.Profile.Fields
			r := csvRow{
				rowType:      "Identity",
				lastName:     profile["last_name"].GetStringValue(),
				firstName:    profile["first_name"].GetStringValue(),
				userID:       profile["user_id"].GetStringValue(),
				userStatus:   getUserStatus(ctx, ut),
				emailAddress: emailAddress,
			}

			err = w.Write(r.Row())
			if err != nil {
				return err
			}
		}
	}

	for _, e := range d.entitlementsByID {
		r := csvRow{
			rowType:                "Entitlement",
			entitlementDisplayName: e.DisplayName,
			entitlement:            e.Id,
			resourceType:           e.Resource.Id.ResourceType,
			resourceName:           e.Resource.DisplayName,
			entitlementDescription: e.Description,
			entitlementSlug:        e.Slug,
		}

		err = w.Write(r.Row())
		if err != nil {
			return err
		}
	}

	for _, g := range d.grantsByID {
		if p, ok := d.resourcesByID[fmtResourceID(g.Principal.Id)]; ok {
			ut := &v2.UserTrait{}

			annos := annotations.Annotations(p.Annotations)
			ok, err = annos.Pick(ut)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}

			var emailAddress string
			for _, e := range ut.Emails {
				if e.IsPrimary {
					emailAddress = e.Address
					break
				}
			}
			profile := ut.Profile.Fields

			var e *v2.Entitlement
			if en, ok := d.entitlementsByID[g.Entitlement.Id]; ok {
				e = en
			} else {
				e = g.Entitlement
			}

			r := csvRow{
				rowType:                "Grant",
				lastName:               profile["last_name"].GetStringValue(),
				firstName:              profile["first_name"].GetStringValue(),
				userID:                 profile["user_id"].GetStringValue(),
				emailAddress:           emailAddress,
				entitlementDisplayName: e.DisplayName,
				entitlement:            e.Id,
				resourceType:           e.Resource.Id.ResourceType,
				resourceName:           e.Resource.DisplayName,
				entitlementDescription: e.Description,
				entitlementSlug:        e.Slug,
			}

			err = w.Write(r.Row())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type dataBag struct {
	resourceTypes      map[string]*v2.ResourceType
	resourcesByID      map[string]*v2.Resource
	resourcesByType    map[string]map[string]*v2.Resource
	entitlementsByID   map[string]*v2.Entitlement
	entitlementsByType map[string]map[string]*v2.Entitlement
	grantsByID         map[string]*v2.Grant
	grantsByType       map[string]map[string]*v2.Grant
}

func fmtResourceID(r *v2.ResourceId) string {
	return fmt.Sprintf("%s:%s", r.ResourceType, r.Resource)
}
