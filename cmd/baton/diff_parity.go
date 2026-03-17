package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/connectorstore"
	"github.com/conductorone/baton-sdk/pkg/dotc1z"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type parityScope string

const (
	parityScopeAll         parityScope = "all"
	parityScopeResource    parityScope = "resource"
	parityScopeEntitlement parityScope = "entitlement"
	parityScopeGrant       parityScope = "grant"
	unknownResourceType                = "<unknown>"
)

type parityOptions struct {
	Scope         parityScope
	ResourceTypes map[string]struct{}
}

type paritySnapshot struct {
	Resources       map[string]*parityResourceRecord
	ResourceRefs    map[string]struct{}
	Entitlements    map[string]*parityEntitlementRecord
	EntitlementRefs map[string]struct{}
	Grants          map[string]*parityGrantRecord
	DanglingGrants  []parityDanglingReference
}

type parityDiffOutput struct {
	Summary               parityDiffSummary         `json:"summary"`
	GroupedByFamily       parityGroupedFamilyDiffs  `json:"groupedByFamily"`
	GroupedByResourceType []parityResourceTypeGroup `json:"groupedByResourceType"`
	LikelyCauses          []parityLikelyCause       `json:"likelyCauses,omitempty"`
	Resources             parityResourceResults     `json:"resources"`
	Entitlements          parityEntitlementResults  `json:"entitlements"`
	Grants                parityGrantResults        `json:"grants"`
}

type parityDiffSummary struct {
	ResourcesOnlyInLeft        int `json:"resourcesOnlyInLeft"`
	ResourcesOnlyInRight       int `json:"resourcesOnlyInRight"`
	ResourceFieldMismatches    int `json:"resourceFieldMismatches"`
	EntitlementsOnlyInLeft     int `json:"entitlementsOnlyInLeft"`
	EntitlementsOnlyInRight    int `json:"entitlementsOnlyInRight"`
	EntitlementFieldMismatches int `json:"entitlementFieldMismatches"`
	GrantsOnlyInLeft           int `json:"grantsOnlyInLeft"`
	GrantsOnlyInRight          int `json:"grantsOnlyInRight"`
	GrantFieldMismatches       int `json:"grantFieldMismatches"`
	DanglingReferencesInLeft   int `json:"danglingReferencesInLeft"`
	DanglingReferencesInRight  int `json:"danglingReferencesInRight"`
}

type parityResourceResults struct {
	OnlyInLeft      []*parityResourceRecord `json:"onlyInLeft"`
	OnlyInRight     []*parityResourceRecord `json:"onlyInRight"`
	FieldMismatches []parityMismatch        `json:"fieldMismatches"`
}

type parityEntitlementResults struct {
	OnlyInLeft      []*parityEntitlementRecord `json:"onlyInLeft"`
	OnlyInRight     []*parityEntitlementRecord `json:"onlyInRight"`
	FieldMismatches []parityMismatch           `json:"fieldMismatches"`
}

type parityGrantResults struct {
	OnlyInLeft         []*parityGrantRecord     `json:"onlyInLeft"`
	OnlyInRight        []*parityGrantRecord     `json:"onlyInRight"`
	FieldMismatches    []parityMismatch         `json:"fieldMismatches"`
	DanglingReferences parityDanglingReferences `json:"danglingReferences"`
}

type parityGroupedFamilyDiffs struct {
	MissingResources    []paritySideGroup     `json:"missingResources"`
	MissingEntitlements []paritySideGroup     `json:"missingEntitlements"`
	MissingGrants       []paritySideGroup     `json:"missingGrants"`
	FieldMismatches     []parityMismatchGroup `json:"fieldMismatches"`
	DanglingReferences  []parityDanglingGroup `json:"danglingReferences"`
}

type paritySideGroup struct {
	ResourceTypeID string   `json:"resourceTypeId"`
	LeftCount      int      `json:"leftCount"`
	RightCount     int      `json:"rightCount"`
	LeftKeys       []string `json:"leftKeys,omitempty"`
	RightKeys      []string `json:"rightKeys,omitempty"`
}

type parityMismatchGroup struct {
	ResourceTypeID string   `json:"resourceTypeId"`
	Family         string   `json:"family"`
	Count          int      `json:"count"`
	Keys           []string `json:"keys,omitempty"`
	Fields         []string `json:"fields,omitempty"`
}

type parityDanglingGroup struct {
	ResourceTypeID string              `json:"resourceTypeId"`
	LeftCount      int                 `json:"leftCount"`
	RightCount     int                 `json:"rightCount"`
	Reasons        []parityReasonCount `json:"reasons,omitempty"`
	LeftExamples   []string            `json:"leftExamples,omitempty"`
	RightExamples  []string            `json:"rightExamples,omitempty"`
}

type parityReasonCount struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type parityResourceTypeGroup struct {
	ResourceTypeID          string                        `json:"resourceTypeId"`
	MissingInventory        paritySideCount               `json:"missingInventory"`
	MissingEntitlements     paritySideCount               `json:"missingEntitlements"`
	MissingGrants           paritySideCount               `json:"missingGrants"`
	FieldMismatches         []parityFamilyCount           `json:"fieldMismatches,omitempty"`
	PrincipalTypeMismatches []parityPrincipalTypeMismatch `json:"principalTypeMismatches,omitempty"`
	DanglingReferences      []parityReasonSideCount       `json:"danglingReferences,omitempty"`
	LikelyCauses            []string                      `json:"likelyCauses,omitempty"`
}

type paritySideCount struct {
	LeftCount  int `json:"leftCount"`
	RightCount int `json:"rightCount"`
}

type parityFamilyCount struct {
	Family string `json:"family"`
	Count  int    `json:"count"`
}

type parityReasonSideCount struct {
	Reason     string `json:"reason"`
	LeftCount  int    `json:"leftCount"`
	RightCount int    `json:"rightCount"`
}

type parityPrincipalTypeMismatch struct {
	ResourceTypeID     string   `json:"resourceTypeId"`
	LeftPrincipalType  string   `json:"leftPrincipalType"`
	RightPrincipalType string   `json:"rightPrincipalType"`
	Count              int      `json:"count"`
	ExampleGrantKeys   []string `json:"exampleGrantKeys,omitempty"`
}

type parityLikelyCause struct {
	ResourceTypeID string `json:"resourceTypeId,omitempty"`
	Category       string `json:"category"`
	Message        string `json:"message"`
}

type parityDanglingReferences struct {
	Left  []parityDanglingReference `json:"left"`
	Right []parityDanglingReference `json:"right"`
}

type parityMismatch struct {
	Family          string                  `json:"family"`
	ResourceTypeID  string                  `json:"resourceTypeId,omitempty"`
	PrincipalTypeID string                  `json:"principalTypeId,omitempty"`
	Key             string                  `json:"key"`
	Differences     []parityFieldDifference `json:"differences"`
}

type parityFieldDifference struct {
	Field string `json:"field"`
	Left  any    `json:"left"`
	Right any    `json:"right"`
}

type parityDanglingReference struct {
	GrantKey        string           `json:"grantKey"`
	ResourceTypeID  string           `json:"resourceTypeId,omitempty"`
	PrincipalTypeID string           `json:"principalTypeId,omitempty"`
	Tuple           parityGrantTuple `json:"tuple"`
	Reason          string           `json:"reason"`
	Reference       string           `json:"reference,omitempty"`
	Value           any              `json:"value"`
}

type parityResourceKey struct {
	ResourceTypeID string `json:"resourceTypeId"`
	ResourceID     string `json:"resourceId"`
}

type parityResourceRecord struct {
	Key              parityResourceKey  `json:"key"`
	DisplayName      string             `json:"displayName"`
	Description      string             `json:"description"`
	ParentResourceID *parityResourceKey `json:"parentResourceId"`
	Annotations      map[string]any     `json:"annotations"`
}

type parityResourceTypeRecord struct {
	ID                string         `json:"id"`
	DisplayName       string         `json:"displayName"`
	Description       string         `json:"description"`
	Traits            []string       `json:"traits"`
	SourcedExternally bool           `json:"sourcedExternally"`
	Annotations       map[string]any `json:"annotations"`
}

type parityEntitlementRecord struct {
	ID          string                     `json:"id"`
	Resource    *parityResourceKey         `json:"resource"`
	DisplayName string                     `json:"displayName"`
	Description string                     `json:"description"`
	Slug        string                     `json:"slug"`
	Purpose     string                     `json:"purpose"`
	GrantableTo []parityResourceTypeRecord `json:"grantableTo"`
	Annotations map[string]any             `json:"annotations"`
}

type parityGrantTuple struct {
	Resource      *parityResourceKey `json:"resource"`
	EntitlementID string             `json:"entitlementId"`
	Principal     *parityResourceKey `json:"principal"`
}

type parityGrantRecord struct {
	Key         string           `json:"key"`
	ID          string           `json:"id"`
	Tuple       parityGrantTuple `json:"tuple"`
	Annotations map[string]any   `json:"annotations"`
}

type parityComparableRecord interface {
	comparisonValue() (any, error)
	sortKey() string
}

type parityMissingValue struct{}

var (
	loadParitySnapshotFn     = loadParitySnapshot
	renderParityDiffOutputFn = renderParityDiffOutput
)

func runParityDiff(cmd *cobra.Command, args []string) error {
	ctx, err := logging.Init(context.Background(), logging.WithLogFormat("console"), logging.WithLogLevel("error"))
	if err != nil {
		return err
	}

	leftPath, err := cmd.Flags().GetString("left")
	if err != nil {
		return err
	}

	rightPath, err := cmd.Flags().GetString("right")
	if err != nil {
		return err
	}

	scopeValue, err := cmd.Flags().GetString("scope")
	if err != nil {
		return err
	}

	scope, err := parseParityScope(scopeValue)
	if err != nil {
		return err
	}

	resourceTypes, err := parseResourceTypes(cmd)
	if err != nil {
		return err
	}

	outputFormat, err := cmd.Flags().GetString("output-format")
	if err != nil {
		return err
	}

	options := parityOptions{
		Scope:         scope,
		ResourceTypes: resourceTypes,
	}

	leftSnapshot, err := loadParitySnapshotFn(ctx, leftPath, options)
	if err != nil {
		return fmt.Errorf("load left snapshot: %w", err)
	}

	rightSnapshot, err := loadParitySnapshotFn(ctx, rightPath, options)
	if err != nil {
		return fmt.Errorf("load right snapshot: %w", err)
	}

	out, err := compareParitySnapshots(leftSnapshot, rightSnapshot)
	if err != nil {
		return err
	}

	return renderParityDiffOutputFn(outputFormat, out)
}

func parseParityScope(value string) (parityScope, error) {
	switch parityScope(strings.TrimSpace(value)) {
	case parityScopeAll:
		return parityScopeAll, nil
	case parityScopeResource:
		return parityScopeResource, nil
	case parityScopeEntitlement:
		return parityScopeEntitlement, nil
	case parityScopeGrant:
		return parityScopeGrant, nil
	default:
		return "", fmt.Errorf("invalid --scope %q: must be one of resource, entitlement, grant, all", value)
	}
}

func parseResourceTypes(cmd *cobra.Command) (map[string]struct{}, error) {
	value, err := cmd.Flags().GetString(resourceTypeFlag)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	ret := make(map[string]struct{})
	for _, part := range strings.Split(value, ",") {
		resourceType := strings.TrimSpace(part)
		if resourceType == "" {
			continue
		}
		ret[resourceType] = struct{}{}
	}

	return ret, nil
}

func (o parityOptions) includeResources() bool {
	return o.Scope == parityScopeAll || o.Scope == parityScopeResource
}

func (o parityOptions) includeEntitlements() bool {
	return o.Scope == parityScopeAll || o.Scope == parityScopeEntitlement
}

func (o parityOptions) includeGrants() bool {
	return o.Scope == parityScopeAll || o.Scope == parityScopeGrant
}

func (o parityOptions) matchesResourceType(resourceTypeID string) bool {
	if len(o.ResourceTypes) == 0 {
		return true
	}

	_, ok := o.ResourceTypes[resourceTypeID]
	return ok
}

func loadParitySnapshot(ctx context.Context, c1zPath string, options parityOptions) (*paritySnapshot, error) {
	m, err := manager.New(ctx, c1zPath)
	if err != nil {
		return nil, err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return nil, err
	}
	defer store.Close(ctx)

	syncID, err := store.LatestSyncID(ctx, connectorstore.SyncTypeFull)
	if err != nil {
		return nil, err
	}

	if syncID == "" {
		return nil, fmt.Errorf("no full syncs found in %s", c1zPath)
	}

	if err := store.ViewSync(ctx, syncID); err != nil {
		return nil, err
	}

	snapshot := &paritySnapshot{
		Resources:       make(map[string]*parityResourceRecord),
		ResourceRefs:    make(map[string]struct{}),
		Entitlements:    make(map[string]*parityEntitlementRecord),
		EntitlementRefs: make(map[string]struct{}),
		Grants:          make(map[string]*parityGrantRecord),
	}

	if err := loadParityResources(ctx, store, snapshot, options); err != nil {
		return nil, err
	}

	if err := loadParityEntitlements(ctx, store, snapshot, options); err != nil {
		return nil, err
	}

	if options.includeGrants() {
		if err := loadParityGrants(ctx, store, snapshot, options); err != nil {
			return nil, err
		}
	}

	sort.Slice(snapshot.DanglingGrants, func(i, j int) bool {
		if snapshot.DanglingGrants[i].GrantKey == snapshot.DanglingGrants[j].GrantKey {
			return snapshot.DanglingGrants[i].Reference < snapshot.DanglingGrants[j].Reference
		}
		return snapshot.DanglingGrants[i].GrantKey < snapshot.DanglingGrants[j].GrantKey
	})

	return snapshot, nil
}

func loadParityResources(ctx context.Context, store *dotc1z.C1File, snapshot *paritySnapshot, options parityOptions) error {
	pageToken := ""
	for {
		resp, err := store.ListResources(ctx, &v2.ResourcesServiceListResourcesRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return err
		}

		for _, resource := range resp.List {
			resourceKey := resourceIDString(resource.GetId())
			snapshot.ResourceRefs[resourceKey] = struct{}{}

			if !options.includeResources() || !options.matchesResourceType(resource.GetId().GetResourceType()) {
				continue
			}

			record, err := newParityResourceRecord(resource)
			if err != nil {
				return err
			}

			snapshot.Resources[record.sortKey()] = record
		}

		if resp.NextPageToken == "" {
			return nil
		}

		pageToken = resp.NextPageToken
	}
}

func loadParityEntitlements(ctx context.Context, store *dotc1z.C1File, snapshot *paritySnapshot, options parityOptions) error {
	pageToken := ""
	for {
		resp, err := store.ListEntitlements(ctx, &v2.EntitlementsServiceListEntitlementsRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return err
		}

		for _, entitlement := range resp.List {
			snapshot.EntitlementRefs[entitlement.GetId()] = struct{}{}

			if !options.includeEntitlements() || !options.matchesResourceType(entitlement.GetResource().GetId().GetResourceType()) {
				continue
			}

			record, err := newParityEntitlementRecord(entitlement)
			if err != nil {
				return err
			}

			snapshot.Entitlements[record.sortKey()] = record
		}

		if resp.NextPageToken == "" {
			return nil
		}

		pageToken = resp.NextPageToken
	}
}

func loadParityGrants(ctx context.Context, store *dotc1z.C1File, snapshot *paritySnapshot, options parityOptions) error {
	pageToken := ""
	for {
		resp, err := store.ListGrants(ctx, &v2.GrantsServiceListGrantsRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return err
		}

		for _, grant := range resp.List {
			if !options.matchesResourceType(grant.GetEntitlement().GetResource().GetId().GetResourceType()) {
				continue
			}

			record, err := newParityGrantRecord(grant)
			if err != nil {
				return err
			}

			snapshot.Grants[record.sortKey()] = record
			snapshot.DanglingGrants = append(snapshot.DanglingGrants, validateGrantReferences(record, snapshot)...)
		}

		if resp.NextPageToken == "" {
			return nil
		}

		pageToken = resp.NextPageToken
	}
}

func compareParitySnapshots(left *paritySnapshot, right *paritySnapshot) (*parityDiffOutput, error) {
	resourceResults, err := compareResourceRecordSets(left.Resources, right.Resources)
	if err != nil {
		return nil, err
	}

	entitlementResults, err := compareEntitlementRecordSets(left.Entitlements, right.Entitlements)
	if err != nil {
		return nil, err
	}

	grantResults, err := compareGrantRecordSets(left.Grants, right.Grants)
	if err != nil {
		return nil, err
	}
	grantResults.DanglingReferences = parityDanglingReferences{
		Left:  left.DanglingGrants,
		Right: right.DanglingGrants,
	}

	out := &parityDiffOutput{
		Summary: parityDiffSummary{
			ResourcesOnlyInLeft:        len(resourceResults.OnlyInLeft),
			ResourcesOnlyInRight:       len(resourceResults.OnlyInRight),
			ResourceFieldMismatches:    len(resourceResults.FieldMismatches),
			EntitlementsOnlyInLeft:     len(entitlementResults.OnlyInLeft),
			EntitlementsOnlyInRight:    len(entitlementResults.OnlyInRight),
			EntitlementFieldMismatches: len(entitlementResults.FieldMismatches),
			GrantsOnlyInLeft:           len(grantResults.OnlyInLeft),
			GrantsOnlyInRight:          len(grantResults.OnlyInRight),
			GrantFieldMismatches:       len(grantResults.FieldMismatches),
			DanglingReferencesInLeft:   len(left.DanglingGrants),
			DanglingReferencesInRight:  len(right.DanglingGrants),
		},
		Resources:    resourceResults,
		Entitlements: entitlementResults,
		Grants:       grantResults,
	}

	enrichParityDiffOutput(out)

	return out, nil
}

func enrichParityDiffOutput(out *parityDiffOutput) {
	out.GroupedByFamily = buildParityFamilyGroups(out)
	out.GroupedByResourceType, out.LikelyCauses = buildParityResourceTypeGroups(out)
}

func buildParityFamilyGroups(out *parityDiffOutput) parityGroupedFamilyDiffs {
	return parityGroupedFamilyDiffs{
		MissingResources:    buildResourceSideGroups(out.Resources.OnlyInLeft, out.Resources.OnlyInRight),
		MissingEntitlements: buildEntitlementSideGroups(out.Entitlements.OnlyInLeft, out.Entitlements.OnlyInRight),
		MissingGrants:       buildGrantSideGroups(out.Grants.OnlyInLeft, out.Grants.OnlyInRight),
		FieldMismatches:     buildMismatchGroups(allParityMismatches(out)),
		DanglingReferences:  buildDanglingGroups(out.Grants.DanglingReferences.Left, out.Grants.DanglingReferences.Right),
	}
}

func buildParityResourceTypeGroups(out *parityDiffOutput) ([]parityResourceTypeGroup, []parityLikelyCause) {
	typeIndex := make(map[string]*parityResourceTypeGroup)

	ensureGroup := func(resourceTypeID string) *parityResourceTypeGroup {
		if resourceTypeID == "" {
			resourceTypeID = unknownResourceType
		}
		if _, ok := typeIndex[resourceTypeID]; !ok {
			typeIndex[resourceTypeID] = &parityResourceTypeGroup{ResourceTypeID: resourceTypeID}
		}
		return typeIndex[resourceTypeID]
	}

	for _, group := range out.GroupedByFamily.MissingResources {
		entry := ensureGroup(group.ResourceTypeID)
		entry.MissingInventory = paritySideCount{LeftCount: group.LeftCount, RightCount: group.RightCount}
	}
	for _, group := range out.GroupedByFamily.MissingEntitlements {
		entry := ensureGroup(group.ResourceTypeID)
		entry.MissingEntitlements = paritySideCount{LeftCount: group.LeftCount, RightCount: group.RightCount}
	}
	for _, group := range out.GroupedByFamily.MissingGrants {
		entry := ensureGroup(group.ResourceTypeID)
		entry.MissingGrants = paritySideCount{LeftCount: group.LeftCount, RightCount: group.RightCount}
	}
	for _, group := range out.GroupedByFamily.FieldMismatches {
		entry := ensureGroup(group.ResourceTypeID)
		entry.FieldMismatches = append(entry.FieldMismatches, parityFamilyCount{Family: group.Family, Count: group.Count})
	}
	for _, group := range out.GroupedByFamily.DanglingReferences {
		entry := ensureGroup(group.ResourceTypeID)
		for _, reason := range group.Reasons {
			sideCount := parityReasonSideCount{Reason: reason.Reason}
			for _, dangling := range out.Grants.DanglingReferences.Left {
				if dangling.ResourceTypeID == group.ResourceTypeID && dangling.Reason == reason.Reason {
					sideCount.LeftCount++
				}
			}
			for _, dangling := range out.Grants.DanglingReferences.Right {
				if dangling.ResourceTypeID == group.ResourceTypeID && dangling.Reason == reason.Reason {
					sideCount.RightCount++
				}
			}
			entry.DanglingReferences = append(entry.DanglingReferences, sideCount)
		}
	}

	for _, mismatch := range detectPrincipalTypeMismatches(out.Grants.OnlyInLeft, out.Grants.OnlyInRight) {
		entry := ensureGroup(mismatch.ResourceTypeID)
		entry.PrincipalTypeMismatches = append(entry.PrincipalTypeMismatches, mismatch)
	}

	resourceTypeIDs := make([]string, 0, len(typeIndex))
	for resourceTypeID := range typeIndex {
		resourceTypeIDs = append(resourceTypeIDs, resourceTypeID)
	}
	sort.Strings(resourceTypeIDs)

	groups := make([]parityResourceTypeGroup, 0, len(resourceTypeIDs))
	var causes []parityLikelyCause
	for _, resourceTypeID := range resourceTypeIDs {
		group := typeIndex[resourceTypeID]
		sort.Slice(group.FieldMismatches, func(i, j int) bool {
			return group.FieldMismatches[i].Family < group.FieldMismatches[j].Family
		})
		sort.Slice(group.PrincipalTypeMismatches, func(i, j int) bool {
			if group.PrincipalTypeMismatches[i].Count == group.PrincipalTypeMismatches[j].Count {
				return group.PrincipalTypeMismatches[i].LeftPrincipalType < group.PrincipalTypeMismatches[j].LeftPrincipalType
			}
			return group.PrincipalTypeMismatches[i].Count > group.PrincipalTypeMismatches[j].Count
		})
		sort.Slice(group.DanglingReferences, func(i, j int) bool {
			return group.DanglingReferences[i].Reason < group.DanglingReferences[j].Reason
		})

		group.LikelyCauses = parityLikelyCauseMessages(group)
		for _, message := range group.LikelyCauses {
			causes = append(causes, parityLikelyCause{
				ResourceTypeID: group.ResourceTypeID,
				Category:       parityLikelyCauseCategory(message),
				Message:        message,
			})
		}
		groups = append(groups, *group)
	}

	return groups, causes
}

func buildResourceSideGroups(left []*parityResourceRecord, right []*parityResourceRecord) []paritySideGroup {
	byType := make(map[string]struct {
		LeftKeys  []string
		RightKeys []string
	})
	for _, record := range left {
		resourceTypeID := record.Key.ResourceTypeID
		group := byType[resourceTypeID]
		group.LeftKeys = append(group.LeftKeys, record.sortKey())
		byType[resourceTypeID] = group
	}
	for _, record := range right {
		resourceTypeID := record.Key.ResourceTypeID
		group := byType[resourceTypeID]
		group.RightKeys = append(group.RightKeys, record.sortKey())
		byType[resourceTypeID] = group
	}
	return paritySideGroupsFromMap(byType)
}

func buildEntitlementSideGroups(left []*parityEntitlementRecord, right []*parityEntitlementRecord) []paritySideGroup {
	byType := make(map[string]struct {
		LeftKeys  []string
		RightKeys []string
	})
	for _, record := range left {
		resourceTypeID := resourceTypeFromKey(record.Resource)
		group := byType[resourceTypeID]
		group.LeftKeys = append(group.LeftKeys, record.sortKey())
		byType[resourceTypeID] = group
	}
	for _, record := range right {
		resourceTypeID := resourceTypeFromKey(record.Resource)
		group := byType[resourceTypeID]
		group.RightKeys = append(group.RightKeys, record.sortKey())
		byType[resourceTypeID] = group
	}
	return paritySideGroupsFromMap(byType)
}

func buildGrantSideGroups(left []*parityGrantRecord, right []*parityGrantRecord) []paritySideGroup {
	byType := make(map[string]struct {
		LeftKeys  []string
		RightKeys []string
	})
	for _, record := range left {
		resourceTypeID := resourceTypeFromKey(record.Tuple.Resource)
		group := byType[resourceTypeID]
		group.LeftKeys = append(group.LeftKeys, record.sortKey())
		byType[resourceTypeID] = group
	}
	for _, record := range right {
		resourceTypeID := resourceTypeFromKey(record.Tuple.Resource)
		group := byType[resourceTypeID]
		group.RightKeys = append(group.RightKeys, record.sortKey())
		byType[resourceTypeID] = group
	}
	return paritySideGroupsFromMap(byType)
}

func paritySideGroupsFromMap(byType map[string]struct {
	LeftKeys  []string
	RightKeys []string
}) []paritySideGroup {
	resourceTypeIDs := make([]string, 0, len(byType))
	for resourceTypeID := range byType {
		resourceTypeIDs = append(resourceTypeIDs, resourceTypeID)
	}
	sort.Strings(resourceTypeIDs)

	var groups []paritySideGroup
	for _, resourceTypeID := range resourceTypeIDs {
		group := byType[resourceTypeID]
		sort.Strings(group.LeftKeys)
		sort.Strings(group.RightKeys)
		groups = append(groups, paritySideGroup{
			ResourceTypeID: resourceTypeID,
			LeftCount:      len(group.LeftKeys),
			RightCount:     len(group.RightKeys),
			LeftKeys:       append([]string(nil), group.LeftKeys...),
			RightKeys:      append([]string(nil), group.RightKeys...),
		})
	}
	return groups
}

func buildMismatchGroups(mismatches []parityMismatch) []parityMismatchGroup {
	type grouped struct {
		Keys   []string
		Fields map[string]struct{}
	}
	byGroup := make(map[string]*grouped)
	groupMeta := make(map[string]parityMismatchGroup)
	for _, mismatch := range mismatches {
		groupKey := mismatch.Family + ":" + mismatch.ResourceTypeID
		if _, ok := byGroup[groupKey]; !ok {
			byGroup[groupKey] = &grouped{Fields: make(map[string]struct{})}
			groupMeta[groupKey] = parityMismatchGroup{
				ResourceTypeID: mismatch.ResourceTypeID,
				Family:         mismatch.Family,
			}
		}
		byGroup[groupKey].Keys = append(byGroup[groupKey].Keys, mismatch.Key)
		for _, difference := range mismatch.Differences {
			byGroup[groupKey].Fields[difference.Field] = struct{}{}
		}
	}

	groupKeys := make([]string, 0, len(byGroup))
	for key := range byGroup {
		groupKeys = append(groupKeys, key)
	}
	sort.Strings(groupKeys)

	var groups []parityMismatchGroup
	for _, key := range groupKeys {
		group := byGroup[key]
		meta := groupMeta[key]
		sort.Strings(group.Keys)
		fields := make([]string, 0, len(group.Fields))
		for field := range group.Fields {
			fields = append(fields, field)
		}
		sort.Strings(fields)
		meta.Count = len(group.Keys)
		meta.Keys = append([]string(nil), group.Keys...)
		meta.Fields = fields
		groups = append(groups, meta)
	}
	return groups
}

func buildDanglingGroups(left []parityDanglingReference, right []parityDanglingReference) []parityDanglingGroup {
	type grouped struct {
		leftReasons   map[string]int
		rightReasons  map[string]int
		leftExamples  []string
		rightExamples []string
	}
	byType := make(map[string]*grouped)
	add := func(values []parityDanglingReference, leftSide bool) {
		for _, value := range values {
			resourceTypeID := value.ResourceTypeID
			if resourceTypeID == "" {
				resourceTypeID = "<unknown>"
			}
			if _, ok := byType[resourceTypeID]; !ok {
				byType[resourceTypeID] = &grouped{
					leftReasons:  make(map[string]int),
					rightReasons: make(map[string]int),
				}
			}
			group := byType[resourceTypeID]
			example := fmt.Sprintf("%s:%s", value.GrantKey, value.Reason)
			if leftSide {
				group.leftReasons[value.Reason]++
				group.leftExamples = append(group.leftExamples, example)
			} else {
				group.rightReasons[value.Reason]++
				group.rightExamples = append(group.rightExamples, example)
			}
		}
	}
	add(left, true)
	add(right, false)

	resourceTypeIDs := make([]string, 0, len(byType))
	for resourceTypeID := range byType {
		resourceTypeIDs = append(resourceTypeIDs, resourceTypeID)
	}
	sort.Strings(resourceTypeIDs)

	var groups []parityDanglingGroup
	for _, resourceTypeID := range resourceTypeIDs {
		group := byType[resourceTypeID]
		sort.Strings(group.leftExamples)
		sort.Strings(group.rightExamples)
		reasonSet := make(map[string]struct{})
		for reason := range group.leftReasons {
			reasonSet[reason] = struct{}{}
		}
		for reason := range group.rightReasons {
			reasonSet[reason] = struct{}{}
		}
		reasons := make([]string, 0, len(reasonSet))
		for reason := range reasonSet {
			reasons = append(reasons, reason)
		}
		sort.Strings(reasons)

		reasonCounts := make([]parityReasonCount, 0, len(reasons))
		leftCount := 0
		rightCount := 0
		for _, reason := range reasons {
			count := group.leftReasons[reason] + group.rightReasons[reason]
			leftCount += group.leftReasons[reason]
			rightCount += group.rightReasons[reason]
			reasonCounts = append(reasonCounts, parityReasonCount{
				Reason: reason,
				Count:  count,
			})
		}

		groups = append(groups, parityDanglingGroup{
			ResourceTypeID: resourceTypeID,
			LeftCount:      leftCount,
			RightCount:     rightCount,
			Reasons:        reasonCounts,
			LeftExamples:   append([]string(nil), takeFirst(group.leftExamples, 5)...),
			RightExamples:  append([]string(nil), takeFirst(group.rightExamples, 5)...),
		})
	}
	return groups
}

func detectPrincipalTypeMismatches(left []*parityGrantRecord, right []*parityGrantRecord) []parityPrincipalTypeMismatch {
	type recordBucket struct {
		Left  []*parityGrantRecord
		Right []*parityGrantRecord
	}
	type mismatchBucket struct {
		Count    int
		Examples []string
	}

	indexKey := func(record *parityGrantRecord) string {
		principalID := ""
		if record.Tuple.Principal != nil {
			principalID = record.Tuple.Principal.ResourceID
		}
		return fmt.Sprintf("%s|%s|%s", resourceKeyString(record.Tuple.Resource), record.Tuple.EntitlementID, principalID)
	}

	grouped := make(map[string]*recordBucket)
	for _, record := range left {
		key := indexKey(record)
		if _, ok := grouped[key]; !ok {
			grouped[key] = &recordBucket{}
		}
		grouped[key].Left = append(grouped[key].Left, record)
	}
	for _, record := range right {
		key := indexKey(record)
		if _, ok := grouped[key]; !ok {
			grouped[key] = &recordBucket{}
		}
		grouped[key].Right = append(grouped[key].Right, record)
	}

	mismatchGroups := make(map[string]*mismatchBucket)
	for _, bucket := range grouped {
		if len(bucket.Left) == 0 || len(bucket.Right) == 0 {
			continue
		}
		leftCounts := principalTypeCounts(bucket.Left)
		rightCounts := principalTypeCounts(bucket.Right)
		if len(leftCounts) != 1 || len(rightCounts) != 1 {
			continue
		}

		leftType := singlePrincipalType(leftCounts)
		rightType := singlePrincipalType(rightCounts)
		if leftType == rightType {
			continue
		}

		groupKey := strings.Join([]string{
			resourceTypeFromKey(bucket.Left[0].Tuple.Resource),
			leftType,
			rightType,
		}, "|")
		if _, ok := mismatchGroups[groupKey]; !ok {
			mismatchGroups[groupKey] = &mismatchBucket{}
		}
		count := len(bucket.Left)
		if len(bucket.Right) < count {
			count = len(bucket.Right)
		}
		mismatchGroups[groupKey].Count += count
		examples := make([]string, 0, len(bucket.Left)+len(bucket.Right))
		for _, record := range bucket.Left {
			examples = append(examples, record.Key)
		}
		for _, record := range bucket.Right {
			examples = append(examples, record.Key)
		}
		sort.Strings(examples)
		mismatchGroups[groupKey].Examples = append(mismatchGroups[groupKey].Examples, takeFirst(examples, 2)...)
	}

	keys := make([]string, 0, len(mismatchGroups))
	for key := range mismatchGroups {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	ret := make([]parityPrincipalTypeMismatch, 0, len(keys))
	for _, key := range keys {
		parts := strings.Split(key, "|")
		bucket := mismatchGroups[key]
		sort.Strings(bucket.Examples)
		ret = append(ret, parityPrincipalTypeMismatch{
			ResourceTypeID:     parts[0],
			LeftPrincipalType:  parts[1],
			RightPrincipalType: parts[2],
			Count:              bucket.Count,
			ExampleGrantKeys:   append([]string(nil), takeFirst(bucket.Examples, 5)...),
		})
	}
	return ret
}

func principalTypeCounts(records []*parityGrantRecord) map[string]int {
	ret := make(map[string]int)
	for _, record := range records {
		ret[resourceTypeFromKey(record.Tuple.Principal)]++
	}
	return ret
}

func singlePrincipalType(counts map[string]int) string {
	for principalType := range counts {
		return principalType
	}
	return "<unknown>"
}

func parityLikelyCauseMessages(group *parityResourceTypeGroup) []string {
	var messages []string

	if group.MissingInventory.RightCount >= 2 && group.MissingInventory.RightCount >= group.MissingInventory.LeftCount {
		messages = append(messages, fmt.Sprintf("many resources only in right for %s suggests upstream inventory gap or pagination issue", group.ResourceTypeID))
	}
	if group.MissingInventory.LeftCount >= 2 && group.MissingInventory.LeftCount > group.MissingInventory.RightCount {
		messages = append(messages, fmt.Sprintf("many resources only in left for %s suggests upstream inventory gap or pagination issue", group.ResourceTypeID))
	}
	for _, mismatch := range group.PrincipalTypeMismatches {
		if mismatch.Count >= 2 {
			messages = append(
				messages,
				fmt.Sprintf(
					"many grant mismatches with %s principals against %s for %s suggest principal classification mismatch",
					mismatch.LeftPrincipalType,
					mismatch.RightPrincipalType,
					group.ResourceTypeID,
				),
			)
		}
	}
	for _, dangling := range group.DanglingReferences {
		if dangling.Reason == "missing_entitlement" && dangling.LeftCount+dangling.RightCount >= 2 {
			messages = append(messages, fmt.Sprintf("many dangling entitlement refs for %s suggest entitlements not emitted for IDs referenced by grants", group.ResourceTypeID))
		}
	}
	return messages
}

func parityLikelyCauseCategory(message string) string {
	switch {
	case strings.Contains(message, "inventory gap or pagination issue"):
		return "inventory_gap"
	case strings.Contains(message, "principal classification mismatch"):
		return "principal_type_mismatch"
	case strings.Contains(message, "entitlements not emitted"):
		return "dangling_entitlement_refs"
	default:
		return "triage_hint"
	}
}

func allParityMismatches(out *parityDiffOutput) []parityMismatch {
	ret := make([]parityMismatch, 0, len(out.Resources.FieldMismatches)+len(out.Entitlements.FieldMismatches)+len(out.Grants.FieldMismatches))
	ret = append(ret, out.Resources.FieldMismatches...)
	ret = append(ret, out.Entitlements.FieldMismatches...)
	ret = append(ret, out.Grants.FieldMismatches...)
	sortParityMismatches(ret)
	return ret
}

func takeFirst(values []string, count int) []string {
	if len(values) <= count {
		return values
	}
	return values[:count]
}

func takeFirstDifferences(values []parityFieldDifference, count int) []parityFieldDifference {
	if len(values) <= count {
		return values
	}
	return values[:count]
}

func compareResourceRecordSets(left map[string]*parityResourceRecord, right map[string]*parityResourceRecord) (parityResourceResults, error) {
	results := parityResourceResults{}

	for key, leftRecord := range left {
		rightRecord, ok := right[key]
		if !ok {
			results.OnlyInLeft = append(results.OnlyInLeft, leftRecord)
			continue
		}

		differences, err := compareComparableRecords(leftRecord, rightRecord)
		if err != nil {
			return parityResourceResults{}, err
		}
		if len(differences) > 0 {
			results.FieldMismatches = append(results.FieldMismatches, parityMismatch{
				Family:         "resource",
				ResourceTypeID: leftRecord.Key.ResourceTypeID,
				Key:            key,
				Differences:    differences,
			})
		}
	}

	for key, rightRecord := range right {
		if _, ok := left[key]; !ok {
			results.OnlyInRight = append(results.OnlyInRight, rightRecord)
		}
	}

	sortParityResources(results.OnlyInLeft)
	sortParityResources(results.OnlyInRight)
	sortParityMismatches(results.FieldMismatches)

	return results, nil
}

func compareEntitlementRecordSets(left map[string]*parityEntitlementRecord, right map[string]*parityEntitlementRecord) (parityEntitlementResults, error) {
	results := parityEntitlementResults{}

	for key, leftRecord := range left {
		rightRecord, ok := right[key]
		if !ok {
			results.OnlyInLeft = append(results.OnlyInLeft, leftRecord)
			continue
		}

		differences, err := compareComparableRecords(leftRecord, rightRecord)
		if err != nil {
			return parityEntitlementResults{}, err
		}
		if len(differences) > 0 {
			results.FieldMismatches = append(results.FieldMismatches, parityMismatch{
				Family:         "entitlement",
				ResourceTypeID: resourceTypeFromKey(leftRecord.Resource),
				Key:            key,
				Differences:    differences,
			})
		}
	}

	for key, rightRecord := range right {
		if _, ok := left[key]; !ok {
			results.OnlyInRight = append(results.OnlyInRight, rightRecord)
		}
	}

	sortParityEntitlements(results.OnlyInLeft)
	sortParityEntitlements(results.OnlyInRight)
	sortParityMismatches(results.FieldMismatches)

	return results, nil
}

func compareGrantRecordSets(left map[string]*parityGrantRecord, right map[string]*parityGrantRecord) (parityGrantResults, error) {
	results := parityGrantResults{}

	leftOnly := make(map[string]*parityGrantRecord)
	rightOnly := make(map[string]*parityGrantRecord)

	for key, leftRecord := range left {
		rightRecord, ok := right[key]
		if !ok {
			leftOnly[key] = leftRecord
			continue
		}

		differences, err := compareComparableRecords(leftRecord, rightRecord)
		if err != nil {
			return parityGrantResults{}, err
		}
		if len(differences) > 0 {
			results.FieldMismatches = append(results.FieldMismatches, parityMismatch{
				Family:          "grant",
				ResourceTypeID:  resourceTypeFromKey(leftRecord.Tuple.Resource),
				PrincipalTypeID: resourceTypeFromKey(leftRecord.Tuple.Principal),
				Key:             key,
				Differences:     differences,
			})
		}
	}

	for key, rightRecord := range right {
		if _, ok := left[key]; !ok {
			rightOnly[key] = rightRecord
		}
	}

	leftByTuple := make(map[string]*parityGrantRecord)
	for key, record := range leftOnly {
		tupleKey := record.Tuple.String()
		if _, ok := leftByTuple[tupleKey]; !ok {
			leftByTuple[tupleKey] = record
		} else {
			results.OnlyInLeft = append(results.OnlyInLeft, leftOnly[key])
		}
	}

	rightByTuple := make(map[string]*parityGrantRecord)
	for key, record := range rightOnly {
		tupleKey := record.Tuple.String()
		if _, ok := rightByTuple[tupleKey]; !ok {
			rightByTuple[tupleKey] = record
		} else {
			results.OnlyInRight = append(results.OnlyInRight, rightOnly[key])
		}
	}

	for tupleKey, leftRecord := range leftByTuple {
		if rightRecord, ok := rightByTuple[tupleKey]; ok {
			differences, err := compareComparableRecords(leftRecord, rightRecord)
			if err != nil {
				return parityGrantResults{}, err
			}
			if len(differences) > 0 {
				results.FieldMismatches = append(results.FieldMismatches, parityMismatch{
					Family:          "grant",
					ResourceTypeID:  resourceTypeFromKey(leftRecord.Tuple.Resource),
					PrincipalTypeID: resourceTypeFromKey(leftRecord.Tuple.Principal),
					Key:             "tuple:" + tupleKey,
					Differences:     differences,
				})
			}
			delete(leftOnly, leftRecord.Key)
			delete(rightOnly, rightRecord.Key)
		}
	}

	for _, leftRecord := range leftOnly {
		results.OnlyInLeft = append(results.OnlyInLeft, leftRecord)
	}
	for _, rightRecord := range rightOnly {
		results.OnlyInRight = append(results.OnlyInRight, rightRecord)
	}

	sortParityGrants(results.OnlyInLeft)
	sortParityGrants(results.OnlyInRight)
	sortParityMismatches(results.FieldMismatches)

	return results, nil
}

func compareComparableRecords(left parityComparableRecord, right parityComparableRecord) ([]parityFieldDifference, error) {
	leftValue, err := left.comparisonValue()
	if err != nil {
		return nil, err
	}

	rightValue, err := right.comparisonValue()
	if err != nil {
		return nil, err
	}

	return diffNormalizedValues("", leftValue, rightValue), nil
}

func diffNormalizedValues(path string, left any, right any) []parityFieldDifference {
	if _, ok := left.(parityMissingValue); ok {
		return []parityFieldDifference{{
			Field: path,
			Left:  "<missing>",
			Right: printableParityValue(right),
		}}
	}

	if _, ok := right.(parityMissingValue); ok {
		return []parityFieldDifference{{
			Field: path,
			Left:  printableParityValue(left),
			Right: "<missing>",
		}}
	}

	leftMap, leftIsMap := left.(map[string]any)
	rightMap, rightIsMap := right.(map[string]any)
	if leftIsMap || rightIsMap {
		if !leftIsMap || !rightIsMap {
			return []parityFieldDifference{{
				Field: path,
				Left:  printableParityValue(left),
				Right: printableParityValue(right),
			}}
		}

		keys := make(map[string]struct{})
		for key := range leftMap {
			keys[key] = struct{}{}
		}
		for key := range rightMap {
			keys[key] = struct{}{}
		}

		sortedKeys := make([]string, 0, len(keys))
		for key := range keys {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)

		var differences []parityFieldDifference
		for _, key := range sortedKeys {
			leftValue, leftOK := leftMap[key]
			if !leftOK {
				leftValue = parityMissingValue{}
			}

			rightValue, rightOK := rightMap[key]
			if !rightOK {
				rightValue = parityMissingValue{}
			}

			differences = append(differences, diffNormalizedValues(joinParityPath(path, key), leftValue, rightValue)...)
		}

		return differences
	}

	leftSlice, leftIsSlice := left.([]any)
	rightSlice, rightIsSlice := right.([]any)
	if leftIsSlice || rightIsSlice {
		if !leftIsSlice || !rightIsSlice {
			return []parityFieldDifference{{
				Field: path,
				Left:  printableParityValue(left),
				Right: printableParityValue(right),
			}}
		}

		maxLength := len(leftSlice)
		if len(rightSlice) > maxLength {
			maxLength = len(rightSlice)
		}

		var differences []parityFieldDifference
		for i := 0; i < maxLength; i++ {
			leftValue := any(parityMissingValue{})
			if i < len(leftSlice) {
				leftValue = leftSlice[i]
			}

			rightValue := any(parityMissingValue{})
			if i < len(rightSlice) {
				rightValue = rightSlice[i]
			}

			differences = append(differences, diffNormalizedValues(fmt.Sprintf("%s[%d]", path, i), leftValue, rightValue)...)
		}

		return differences
	}

	if canonicalParityJSON(left) == canonicalParityJSON(right) {
		return nil
	}

	return []parityFieldDifference{{
		Field: path,
		Left:  printableParityValue(left),
		Right: printableParityValue(right),
	}}
}

func joinParityPath(path string, key string) string {
	if path == "" {
		return key
	}

	return path + "." + key
}

func printableParityValue(value any) any {
	if _, ok := value.(parityMissingValue); ok {
		return "<missing>"
	}

	return value
}

func newParityResourceRecord(resource *v2.Resource) (*parityResourceRecord, error) {
	annotations, err := normalizeAnnotations(resource.GetAnnotations())
	if err != nil {
		return nil, err
	}

	return &parityResourceRecord{
		Key: parityResourceKey{
			ResourceTypeID: resource.GetId().GetResourceType(),
			ResourceID:     resource.GetId().GetResource(),
		},
		DisplayName:      resource.GetDisplayName(),
		Description:      resource.GetDescription(),
		ParentResourceID: resourceKeyPtr(resource.GetParentResourceId()),
		Annotations:      annotations,
	}, nil
}

func newParityEntitlementRecord(entitlement *v2.Entitlement) (*parityEntitlementRecord, error) {
	grantableTo := make([]parityResourceTypeRecord, 0, len(entitlement.GetGrantableTo()))
	for _, resourceType := range entitlement.GetGrantableTo() {
		record, err := normalizeParityResourceType(resourceType)
		if err != nil {
			return nil, err
		}
		grantableTo = append(grantableTo, record)
	}

	annotations, err := normalizeAnnotations(entitlement.GetAnnotations())
	if err != nil {
		return nil, err
	}

	return &parityEntitlementRecord{
		ID:          entitlement.GetId(),
		Resource:    resourceKeyPtr(entitlement.GetResource().GetId()),
		DisplayName: entitlement.GetDisplayName(),
		Description: entitlement.GetDescription(),
		Slug:        entitlement.GetSlug(),
		Purpose:     entitlement.GetPurpose().String(),
		GrantableTo: grantableTo,
		Annotations: annotations,
	}, nil
}

func newParityGrantRecord(grant *v2.Grant) (*parityGrantRecord, error) {
	annotations, err := normalizeAnnotations(grant.GetAnnotations())
	if err != nil {
		return nil, err
	}

	tuple := parityGrantTuple{
		Resource:      resourceKeyPtr(grant.GetEntitlement().GetResource().GetId()),
		EntitlementID: grant.GetEntitlement().GetId(),
		Principal:     resourceKeyPtr(grant.GetPrincipal().GetId()),
	}

	key := "tuple:" + tuple.String()
	if grant.GetId() != "" {
		key = "id:" + grant.GetId()
	}

	return &parityGrantRecord{
		Key:         key,
		ID:          grant.GetId(),
		Tuple:       tuple,
		Annotations: annotations,
	}, nil
}

func validateGrantReferences(grant *parityGrantRecord, snapshot *paritySnapshot) []parityDanglingReference {
	var dangling []parityDanglingReference

	if _, ok := snapshot.EntitlementRefs[grant.Tuple.EntitlementID]; !ok {
		dangling = append(dangling, parityDanglingReference{
			GrantKey:        grant.Key,
			ResourceTypeID:  resourceTypeFromKey(grant.Tuple.Resource),
			PrincipalTypeID: resourceTypeFromKey(grant.Tuple.Principal),
			Tuple:           grant.Tuple,
			Reason:          "missing_entitlement",
			Reference:       "entitlement",
			Value:           grant.Tuple.EntitlementID,
		})
	}

	resourceKey := resourceKeyString(grant.Tuple.Resource)
	if _, ok := snapshot.ResourceRefs[resourceKey]; !ok {
		dangling = append(dangling, parityDanglingReference{
			GrantKey:        grant.Key,
			ResourceTypeID:  resourceTypeFromKey(grant.Tuple.Resource),
			PrincipalTypeID: resourceTypeFromKey(grant.Tuple.Principal),
			Tuple:           grant.Tuple,
			Reason:          "missing_grant_resource",
			Reference:       "grantResource",
			Value:           grant.Tuple.Resource,
		})
	}

	principalKey := resourceKeyString(grant.Tuple.Principal)
	if _, ok := snapshot.ResourceRefs[principalKey]; !ok {
		dangling = append(dangling, parityDanglingReference{
			GrantKey:        grant.Key,
			ResourceTypeID:  resourceTypeFromKey(grant.Tuple.Resource),
			PrincipalTypeID: resourceTypeFromKey(grant.Tuple.Principal),
			Tuple:           grant.Tuple,
			Reason:          "missing_principal",
			Reference:       "principal",
			Value:           grant.Tuple.Principal,
		})
	}

	return dangling
}

func normalizeAnnotations(values []*anypb.Any) (map[string]any, error) {
	if len(values) == 0 {
		return nil, nil
	}

	ret := make(map[string]any)
	for _, annotationValue := range values {
		key, normalized, err := normalizeAnnotation(annotationValue)
		if err != nil {
			return nil, err
		}
		if existing, ok := ret[key]; ok {
			switch typed := existing.(type) {
			case []any:
				ret[key] = append(typed, normalized)
			default:
				ret[key] = []any{typed, normalized}
			}
			continue
		}
		ret[key] = normalized
	}

	if len(ret) == 0 {
		return nil, nil
	}

	normalized, err := normalizeComparableValue(ret)
	if err != nil {
		return nil, err
	}

	normalizedMap, ok := normalized.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("normalized annotations had unexpected type %T", normalized)
	}

	return normalizedMap, nil
}

func normalizeAnnotation(annotationValue *anypb.Any) (string, any, error) {
	if annotationValue == nil {
		return "<nil>", nil, nil
	}

	message, err := anypb.UnmarshalNew(annotationValue, proto.UnmarshalOptions{})
	if err != nil {
		normalized, marshalErr := normalizeProtoValue(annotationValue)
		if marshalErr != nil {
			return annotationValue.GetTypeUrl(), nil, marshalErr
		}
		return annotationKeyFromTypeURL(annotationValue.GetTypeUrl()), normalized, nil
	}

	normalized, err := normalizeProtoValue(message)
	if err != nil {
		return annotationKeyFromTypeURL(annotationValue.GetTypeUrl()), nil, err
	}

	return annotationKeyForMessage(message, annotationValue.GetTypeUrl()), normalized, nil
}

func annotationKeyForMessage(message proto.Message, fallbackTypeURL string) string {
	if message == nil {
		return annotationKeyFromTypeURL(fallbackTypeURL)
	}
	return lowerFirst(string(message.ProtoReflect().Descriptor().Name()))
}

func annotationKeyFromTypeURL(typeURL string) string {
	if typeURL == "" {
		return "<unknownAnnotation>"
	}
	parts := strings.Split(typeURL, ".")
	return lowerFirst(parts[len(parts)-1])
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}
	return strings.ToLower(value[:1]) + value[1:]
}

func normalizeParityResourceType(resourceType *v2.ResourceType) (parityResourceTypeRecord, error) {
	annotations, err := normalizeAnnotations(resourceType.GetAnnotations())
	if err != nil {
		return parityResourceTypeRecord{}, err
	}

	traits := make([]string, 0, len(resourceType.GetTraits()))
	for _, trait := range resourceType.GetTraits() {
		traits = append(traits, trait.String())
	}
	sort.Strings(traits)

	return parityResourceTypeRecord{
		ID:                resourceType.GetId(),
		DisplayName:       resourceType.GetDisplayName(),
		Description:       resourceType.GetDescription(),
		Traits:            traits,
		SourcedExternally: resourceType.GetSourcedExternally(),
		Annotations:       annotations,
	}, nil
}

func normalizeProtoValue(message proto.Message) (any, error) {
	bytes, err := protojson.Marshal(message)
	if err != nil {
		return nil, err
	}

	var value any
	if err := json.Unmarshal(bytes, &value); err != nil {
		return nil, err
	}

	return normalizeParityValueAtPath("", value), nil
}

func normalizeComparableValue(value any) (any, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var normalized any
	if err := json.Unmarshal(bytes, &normalized); err != nil {
		return nil, err
	}

	return normalizeParityValueAtPath("", normalized), nil
}

func normalizeParityValueAtPath(path string, value any) any {
	switch typed := value.(type) {
	case map[string]any:
		ret := make(map[string]any, len(typed))
		for key, val := range typed {
			ret[key] = normalizeParityValueAtPath(joinParityPath(path, key), val)
		}
		return ret
	case []any:
		ret := make([]any, len(typed))
		for i, val := range typed {
			ret[i] = normalizeParityValueAtPath(path, val)
		}
		if shouldSortRepeatedField(path) {
			sort.Slice(ret, func(i, j int) bool {
				return canonicalParityJSON(ret[i]) < canonicalParityJSON(ret[j])
			})
		}
		return ret
	default:
		return typed
	}
}

func shouldSortRepeatedField(path string) bool {
	if path == "" {
		return false
	}

	field := lastParityPathComponent(path)
	switch field {
	case "grantableTo", "traits", "flags", "loginAliases", "employeeIds", "emails", "entitlementIds", "resourceTypeIds", "group_types", "groupTypes":
		return true
	default:
		return false
	}
}

func lastParityPathComponent(path string) string {
	trimmed := path
	if index := strings.LastIndex(trimmed, "."); index >= 0 {
		trimmed = trimmed[index+1:]
	}
	if index := strings.Index(trimmed, "["); index >= 0 {
		trimmed = trimmed[:index]
	}
	return trimmed
}

func canonicalParityJSON(value any) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%#v", value)
	}
	return string(bytes)
}

func sortParityResources(records []*parityResourceRecord) {
	sort.Slice(records, func(i, j int) bool {
		return records[i].sortKey() < records[j].sortKey()
	})
}

func sortParityEntitlements(records []*parityEntitlementRecord) {
	sort.Slice(records, func(i, j int) bool {
		return records[i].sortKey() < records[j].sortKey()
	})
}

func sortParityGrants(records []*parityGrantRecord) {
	sort.Slice(records, func(i, j int) bool {
		return records[i].sortKey() < records[j].sortKey()
	})
}

func sortParityMismatches(mismatches []parityMismatch) {
	sort.Slice(mismatches, func(i, j int) bool {
		return mismatches[i].Key < mismatches[j].Key
	})
}

func resourceIDString(id *v2.ResourceId) string {
	if id == nil {
		return "<missing-resource-id>"
	}
	return fmt.Sprintf("%s:%s", id.GetResourceType(), id.GetResource())
}

func resourceKeyPtr(id *v2.ResourceId) *parityResourceKey {
	if id == nil {
		return nil
	}

	return &parityResourceKey{
		ResourceTypeID: id.GetResourceType(),
		ResourceID:     id.GetResource(),
	}
}

func resourceKeyString(key *parityResourceKey) string {
	if key == nil {
		return "<missing-resource>"
	}
	return key.String()
}

func resourceTypeFromKey(key *parityResourceKey) string {
	if key == nil {
		return unknownResourceType
	}
	return key.ResourceTypeID
}

func (k parityResourceKey) String() string {
	return fmt.Sprintf("%s:%s", k.ResourceTypeID, k.ResourceID)
}

func (t parityGrantTuple) String() string {
	return fmt.Sprintf("%s|%s|%s", resourceKeyString(t.Resource), t.EntitlementID, resourceKeyString(t.Principal))
}

func (r *parityResourceRecord) sortKey() string {
	return r.Key.String()
}

func (r *parityResourceRecord) comparisonValue() (any, error) {
	return normalizeComparableValue(r)
}

func (e *parityEntitlementRecord) sortKey() string {
	return e.ID
}

func (e *parityEntitlementRecord) comparisonValue() (any, error) {
	return normalizeComparableValue(e)
}

func (g *parityGrantRecord) sortKey() string {
	return g.Key
}

func (g *parityGrantRecord) comparisonValue() (any, error) {
	return normalizeComparableValue(g)
}

func renderParityDiffOutput(outputFormat string, out *parityDiffOutput) error {
	switch outputFormat {
	case "json":
		return renderParityJSON(out)
	default:
		return renderParityConsole(out)
	}
}

func renderParityJSON(out *parityDiffOutput) error {
	bytes, err := marshalParityJSON(out)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, string(bytes))
	return err
}

func marshalParityJSON(out *parityDiffOutput) ([]byte, error) {
	return json.MarshalIndent(out, "", "  ")
}

func renderParityConsole(out *parityDiffOutput) error {
	summaryFormat := strings.Join([]string{
		"Summary",
		"  resources only in left: %d",
		"  resources only in right: %d",
		"  resource field mismatches: %d",
		"  entitlements only in left: %d",
		"  entitlements only in right: %d",
		"  entitlement field mismatches: %d",
		"  grants only in left: %d",
		"  grants only in right: %d",
		"  grant field mismatches: %d",
		"  dangling references in left: %d",
		"  dangling references in right: %d",
		"",
	}, "\n")

	if _, err := fmt.Fprintf(os.Stdout,
		summaryFormat,
		out.Summary.ResourcesOnlyInLeft,
		out.Summary.ResourcesOnlyInRight,
		out.Summary.ResourceFieldMismatches,
		out.Summary.EntitlementsOnlyInLeft,
		out.Summary.EntitlementsOnlyInRight,
		out.Summary.EntitlementFieldMismatches,
		out.Summary.GrantsOnlyInLeft,
		out.Summary.GrantsOnlyInRight,
		out.Summary.GrantFieldMismatches,
		out.Summary.DanglingReferencesInLeft,
		out.Summary.DanglingReferencesInRight,
	); err != nil {
		return err
	}

	writeSectionHeader := func(title string) error {
		_, err := fmt.Fprintf(os.Stdout, "\n%s\n", title)
		return err
	}

	writeSideGroups := func(title string, groups []paritySideGroup) error {
		if len(groups) == 0 {
			return nil
		}
		if err := writeSectionHeader(title); err != nil {
			return err
		}
		for _, group := range groups {
			if _, err := fmt.Fprintf(os.Stdout, "  %s: left=%d right=%d\n", group.ResourceTypeID, group.LeftCount, group.RightCount); err != nil {
				return err
			}
			if len(group.LeftKeys) > 0 {
				if _, err := fmt.Fprintf(os.Stdout, "    left examples: %s\n", strings.Join(takeFirst(group.LeftKeys, 3), ", ")); err != nil {
					return err
				}
			}
			if len(group.RightKeys) > 0 {
				if _, err := fmt.Fprintf(os.Stdout, "    right examples: %s\n", strings.Join(takeFirst(group.RightKeys, 3), ", ")); err != nil {
					return err
				}
			}
		}
		return nil
	}

	writeMismatchGroups := func(title string, groups []parityMismatchGroup, mismatches []parityMismatch) error {
		if len(groups) == 0 {
			return nil
		}
		detailsByGroup := make(map[string][]parityMismatch)
		for _, mismatch := range mismatches {
			key := mismatch.Family + ":" + mismatch.ResourceTypeID
			detailsByGroup[key] = append(detailsByGroup[key], mismatch)
		}
		if err := writeSectionHeader(title); err != nil {
			return err
		}
		for _, group := range groups {
			if _, err := fmt.Fprintf(os.Stdout, "  %s / %s: mismatches=%d fields=%s\n", group.ResourceTypeID, group.Family, group.Count, strings.Join(group.Fields, ", ")); err != nil {
				return err
			}
			groupDetails := detailsByGroup[group.Family+":"+group.ResourceTypeID]
			if len(groupDetails) > 3 {
				groupDetails = groupDetails[:3]
			}
			for _, mismatch := range groupDetails {
				if _, err := fmt.Fprintf(os.Stdout, "    %s\n", mismatch.Key); err != nil {
					return err
				}
				for _, difference := range takeFirstDifferences(mismatch.Differences, 3) {
					if _, err := fmt.Fprintf(os.Stdout, "      %s: left=%s right=%s\n", difference.Field, canonicalParityJSON(difference.Left), canonicalParityJSON(difference.Right)); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	writeDanglingGroups := func(title string, groups []parityDanglingGroup) error {
		if len(groups) == 0 {
			return nil
		}
		if err := writeSectionHeader(title); err != nil {
			return err
		}
		for _, group := range groups {
			reasons := make([]string, 0, len(group.Reasons))
			for _, reason := range group.Reasons {
				reasons = append(reasons, fmt.Sprintf("%s=%d", reason.Reason, reason.Count))
			}
			if _, err := fmt.Fprintf(os.Stdout, "  %s: left=%d right=%d reasons=%s\n", group.ResourceTypeID, group.LeftCount, group.RightCount, strings.Join(reasons, ", ")); err != nil {
				return err
			}
			if len(group.LeftExamples) > 0 {
				if _, err := fmt.Fprintf(os.Stdout, "    left examples: %s\n", strings.Join(group.LeftExamples, ", ")); err != nil {
					return err
				}
			}
			if len(group.RightExamples) > 0 {
				if _, err := fmt.Fprintf(os.Stdout, "    right examples: %s\n", strings.Join(group.RightExamples, ", ")); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if len(out.LikelyCauses) > 0 {
		if err := writeSectionHeader("Likely Causes"); err != nil {
			return err
		}
		for _, cause := range out.LikelyCauses {
			if _, err := fmt.Fprintf(os.Stdout, "  [%s] %s\n", cause.ResourceTypeID, cause.Message); err != nil {
				return err
			}
		}
	}

	for _, group := range out.GroupedByResourceType {
		if group.ResourceTypeID != "enterprise_application" {
			continue
		}
		if err := writeSectionHeader("Enterprise Application Focus"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(os.Stdout, "  missing inventory: left=%d right=%d\n", group.MissingInventory.LeftCount, group.MissingInventory.RightCount); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(os.Stdout, "  missing entitlements: left=%d right=%d\n", group.MissingEntitlements.LeftCount, group.MissingEntitlements.RightCount); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(os.Stdout, "  missing grants: left=%d right=%d\n", group.MissingGrants.LeftCount, group.MissingGrants.RightCount); err != nil {
			return err
		}
		if len(group.PrincipalTypeMismatches) > 0 {
			if _, err := fmt.Fprintf(os.Stdout, "  principal type mismatches:\n"); err != nil {
				return err
			}
			for _, mismatch := range group.PrincipalTypeMismatches {
				if _, err := fmt.Fprintf(os.Stdout, "    %s -> %s: %d\n", mismatch.LeftPrincipalType, mismatch.RightPrincipalType, mismatch.Count); err != nil {
					return err
				}
			}
		}
	}

	if err := writeSideGroups("Missing Resources", out.GroupedByFamily.MissingResources); err != nil {
		return err
	}
	if err := writeSideGroups("Missing Entitlements", out.GroupedByFamily.MissingEntitlements); err != nil {
		return err
	}
	if err := writeSideGroups("Missing Grants", out.GroupedByFamily.MissingGrants); err != nil {
		return err
	}
	if err := writeMismatchGroups("Field Mismatches", out.GroupedByFamily.FieldMismatches, allParityMismatches(out)); err != nil {
		return err
	}
	if err := writeDanglingGroups("Dangling References", out.GroupedByFamily.DanglingReferences); err != nil {
		return err
	}

	return nil
}
