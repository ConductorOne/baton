package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCompareParitySnapshotsIgnoresOrdering(t *testing.T) {
	leftResource, err := newParityResourceRecord(&v2.Resource{
		Id:          &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"},
		DisplayName: "Payroll",
		Description: "Payroll app",
		Annotations: mustAnnotations(t,
			&v2.ExternalLink{Url: "https://example.com/apps/payroll"},
			&v2.UserTrait{
				LoginAliases: []string{"b", "a"},
				Profile:      mustStruct(t, map[string]any{"team": "finance", "active": true}),
			},
			&v2.AppTrait{
				HelpUrl: "https://docs.example.com/payroll",
				Flags:   []v2.AppTrait_AppFlag{v2.AppTrait_APP_FLAG_OIDC, v2.AppTrait_APP_FLAG_SAML},
			},
		),
	})
	if err != nil {
		t.Fatalf("new left resource record: %v", err)
	}

	rightResource, err := newParityResourceRecord(&v2.Resource{
		Id:          &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"},
		DisplayName: "Payroll",
		Description: "Payroll app",
		Annotations: mustAnnotations(t,
			&v2.AppTrait{
				HelpUrl: "https://docs.example.com/payroll",
				Flags:   []v2.AppTrait_AppFlag{v2.AppTrait_APP_FLAG_SAML, v2.AppTrait_APP_FLAG_OIDC},
			},
			&v2.UserTrait{
				LoginAliases: []string{"a", "b"},
				Profile:      mustStruct(t, map[string]any{"active": true, "team": "finance"}),
			},
			&v2.ExternalLink{Url: "https://example.com/apps/payroll"},
		),
	})
	if err != nil {
		t.Fatalf("new right resource record: %v", err)
	}

	leftEntitlement, err := newParityEntitlementRecord(&v2.Entitlement{
		Id:          "ent-1",
		DisplayName: "Admin",
		Description: "Admin access",
		Slug:        "admin",
		Purpose:     v2.Entitlement_PURPOSE_VALUE_ASSIGNMENT,
		Resource: &v2.Resource{
			Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"},
		},
		GrantableTo: []*v2.ResourceType{
			{Id: "user", DisplayName: "User"},
			{Id: "group", DisplayName: "Group"},
		},
	})
	if err != nil {
		t.Fatalf("new left entitlement record: %v", err)
	}

	rightEntitlement, err := newParityEntitlementRecord(&v2.Entitlement{
		Id:          "ent-1",
		DisplayName: "Admin",
		Description: "Admin access",
		Slug:        "admin",
		Purpose:     v2.Entitlement_PURPOSE_VALUE_ASSIGNMENT,
		Resource: &v2.Resource{
			Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"},
		},
		GrantableTo: []*v2.ResourceType{
			{Id: "group", DisplayName: "Group"},
			{Id: "user", DisplayName: "User"},
		},
	})
	if err != nil {
		t.Fatalf("new right entitlement record: %v", err)
	}

	leftGrant, err := newParityGrantRecord(&v2.Grant{
		Id: "grant-1",
		Entitlement: &v2.Entitlement{
			Id:       "ent-1",
			Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"}},
		},
		Principal: &v2.Resource{Id: &v2.ResourceId{ResourceType: "user", Resource: "user-1"}},
		Annotations: mustAnnotations(t, &v2.GrantExpandable{
			EntitlementIds:  []string{"ent-2", "ent-1"},
			ResourceTypeIds: []string{"group", "user"},
		}),
	})
	if err != nil {
		t.Fatalf("new left grant record: %v", err)
	}

	rightGrant, err := newParityGrantRecord(&v2.Grant{
		Id: "grant-1",
		Entitlement: &v2.Entitlement{
			Id:       "ent-1",
			Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"}},
		},
		Principal: &v2.Resource{Id: &v2.ResourceId{ResourceType: "user", Resource: "user-1"}},
		Annotations: mustAnnotations(t, &v2.GrantExpandable{
			EntitlementIds:  []string{"ent-1", "ent-2"},
			ResourceTypeIds: []string{"user", "group"},
		}),
	})
	if err != nil {
		t.Fatalf("new right grant record: %v", err)
	}

	out, err := compareParitySnapshots(
		&paritySnapshot{
			Resources:       map[string]*parityResourceRecord{leftResource.sortKey(): leftResource},
			Entitlements:    map[string]*parityEntitlementRecord{leftEntitlement.sortKey(): leftEntitlement},
			Grants:          map[string]*parityGrantRecord{leftGrant.sortKey(): leftGrant},
			ResourceRefs:    map[string]struct{}{},
			EntitlementRefs: map[string]struct{}{},
		},
		&paritySnapshot{
			Resources:       map[string]*parityResourceRecord{rightResource.sortKey(): rightResource},
			Entitlements:    map[string]*parityEntitlementRecord{rightEntitlement.sortKey(): rightEntitlement},
			Grants:          map[string]*parityGrantRecord{rightGrant.sortKey(): rightGrant},
			ResourceRefs:    map[string]struct{}{},
			EntitlementRefs: map[string]struct{}{},
		},
	)
	if err != nil {
		t.Fatalf("compare parity snapshots: %v", err)
	}

	if got := out.Summary.ResourceFieldMismatches + out.Summary.EntitlementFieldMismatches + out.Summary.GrantFieldMismatches; got != 0 {
		t.Fatalf("expected no field mismatches, got %d", got)
	}
	if got := out.Summary.ResourcesOnlyInLeft +
		out.Summary.ResourcesOnlyInRight +
		out.Summary.EntitlementsOnlyInLeft +
		out.Summary.EntitlementsOnlyInRight +
		out.Summary.GrantsOnlyInLeft +
		out.Summary.GrantsOnlyInRight; got != 0 {
		t.Fatalf("expected no missing records, got %d", got)
	}
}

func TestCompareParitySnapshotsDetectsMissingAndFieldMismatches(t *testing.T) {
	leftResource, err := newParityResourceRecord(&v2.Resource{
		Id:          &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"},
		DisplayName: "Left Name",
	})
	if err != nil {
		t.Fatalf("new left resource record: %v", err)
	}

	rightResource, err := newParityResourceRecord(&v2.Resource{
		Id:          &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"},
		DisplayName: "Right Name",
	})
	if err != nil {
		t.Fatalf("new right resource record: %v", err)
	}

	leftEntitlement, err := newParityEntitlementRecord(&v2.Entitlement{
		Id:          "ent-left-only",
		DisplayName: "Left only",
		Resource:    &v2.Resource{Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"}},
	})
	if err != nil {
		t.Fatalf("new left entitlement record: %v", err)
	}

	rightEntitlement, err := newParityEntitlementRecord(&v2.Entitlement{
		Id:          "ent-right-only",
		DisplayName: "Right only",
		Resource:    &v2.Resource{Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"}},
	})
	if err != nil {
		t.Fatalf("new right entitlement record: %v", err)
	}

	out, err := compareParitySnapshots(
		&paritySnapshot{
			Resources:       map[string]*parityResourceRecord{leftResource.sortKey(): leftResource},
			Entitlements:    map[string]*parityEntitlementRecord{leftEntitlement.sortKey(): leftEntitlement},
			Grants:          map[string]*parityGrantRecord{},
			ResourceRefs:    map[string]struct{}{},
			EntitlementRefs: map[string]struct{}{},
		},
		&paritySnapshot{
			Resources:       map[string]*parityResourceRecord{rightResource.sortKey(): rightResource},
			Entitlements:    map[string]*parityEntitlementRecord{rightEntitlement.sortKey(): rightEntitlement},
			Grants:          map[string]*parityGrantRecord{},
			ResourceRefs:    map[string]struct{}{},
			EntitlementRefs: map[string]struct{}{},
		},
	)
	if err != nil {
		t.Fatalf("compare parity snapshots: %v", err)
	}

	if out.Summary.ResourceFieldMismatches != 1 {
		t.Fatalf("expected 1 resource mismatch, got %d", out.Summary.ResourceFieldMismatches)
	}
	if out.Summary.EntitlementsOnlyInLeft != 1 || out.Summary.EntitlementsOnlyInRight != 1 {
		t.Fatalf("expected one entitlement only on each side, got left=%d right=%d", out.Summary.EntitlementsOnlyInLeft, out.Summary.EntitlementsOnlyInRight)
	}
	if len(out.Resources.FieldMismatches) != 1 {
		t.Fatalf("expected one resource field mismatch, got %d", len(out.Resources.FieldMismatches))
	}
	if !containsField(out.Resources.FieldMismatches[0].Differences, "displayName") {
		t.Fatalf("expected displayName difference, got %#v", out.Resources.FieldMismatches[0].Differences)
	}
}

func TestCompareGrantRecordSetsMatchesOnTupleWhenIDsDiffer(t *testing.T) {
	leftGrant, err := newParityGrantRecord(&v2.Grant{
		Id: "left-id",
		Entitlement: &v2.Entitlement{
			Id:       "ent-1",
			Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"}},
		},
		Principal: &v2.Resource{Id: &v2.ResourceId{ResourceType: "user", Resource: "user-1"}},
	})
	if err != nil {
		t.Fatalf("new left grant record: %v", err)
	}

	rightGrant, err := newParityGrantRecord(&v2.Grant{
		Id: "right-id",
		Entitlement: &v2.Entitlement{
			Id:       "ent-1",
			Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"}},
		},
		Principal: &v2.Resource{Id: &v2.ResourceId{ResourceType: "user", Resource: "user-1"}},
	})
	if err != nil {
		t.Fatalf("new right grant record: %v", err)
	}

	out, err := compareGrantRecordSets(
		map[string]*parityGrantRecord{leftGrant.sortKey(): leftGrant},
		map[string]*parityGrantRecord{rightGrant.sortKey(): rightGrant},
	)
	if err != nil {
		t.Fatalf("compare grant records: %v", err)
	}

	if len(out.OnlyInLeft) != 0 || len(out.OnlyInRight) != 0 {
		t.Fatalf("expected tuple reconciliation, got left=%d right=%d", len(out.OnlyInLeft), len(out.OnlyInRight))
	}
	if len(out.FieldMismatches) != 1 {
		t.Fatalf("expected one reconciled grant mismatch, got %d", len(out.FieldMismatches))
	}
	if !strings.HasPrefix(out.FieldMismatches[0].Key, "tuple:") {
		t.Fatalf("expected tuple mismatch key, got %q", out.FieldMismatches[0].Key)
	}
	if !containsField(out.FieldMismatches[0].Differences, "id") {
		t.Fatalf("expected id difference, got %#v", out.FieldMismatches[0].Differences)
	}
}

func TestValidateGrantReferencesDetectsDangling(t *testing.T) {
	grant := &parityGrantRecord{
		Key: "id:grant-1",
		Tuple: parityGrantTuple{
			Resource:      &parityResourceKey{ResourceTypeID: "enterprise_application", ResourceID: "app-1"},
			EntitlementID: "ent-1",
			Principal:     &parityResourceKey{ResourceTypeID: "user", ResourceID: "user-1"},
		},
	}

	dangling := validateGrantReferences(grant, &paritySnapshot{
		ResourceRefs:    map[string]struct{}{"enterprise_application:app-1": {}},
		EntitlementRefs: map[string]struct{}{},
	})

	if len(dangling) != 2 {
		t.Fatalf("expected 2 dangling references, got %d", len(dangling))
	}
	if dangling[0].Reference != "entitlement" {
		t.Fatalf("expected entitlement dangling reference first, got %q", dangling[0].Reference)
	}
	if dangling[1].Reference != "principal" {
		t.Fatalf("expected principal dangling reference second, got %q", dangling[1].Reference)
	}
}

func TestMarshalParityJSONIsStable(t *testing.T) {
	out := &parityDiffOutput{
		Summary: parityDiffSummary{
			ResourcesOnlyInLeft: 1,
		},
		Resources: parityResourceResults{
			OnlyInLeft: []*parityResourceRecord{{
				Key:         parityResourceKey{ResourceTypeID: "enterprise_application", ResourceID: "app-1"},
				DisplayName: "Payroll",
				Annotations: map[string]any{"b": true, "a": "x"},
			}},
		},
	}

	first, err := marshalParityJSON(out)
	if err != nil {
		t.Fatalf("first marshal: %v", err)
	}

	second, err := marshalParityJSON(out)
	if err != nil {
		t.Fatalf("second marshal: %v", err)
	}

	if string(first) != string(second) {
		t.Fatalf("expected stable json output:\n%s\n!=\n%s", string(first), string(second))
	}
	if !strings.Contains(string(first), "\"resourcesOnlyInLeft\": 1") {
		t.Fatalf("expected summary count in json output, got %s", string(first))
	}
}

func TestRunDiffDispatchesModes(t *testing.T) {
	originalLegacy := legacyDiffRunner
	originalParity := parityDiffRunner
	defer func() {
		legacyDiffRunner = originalLegacy
		parityDiffRunner = originalParity
	}()

	t.Run("legacy default", func(t *testing.T) {
		cmd := diffCmd()

		calledLegacy := false
		calledParity := false
		legacyDiffRunner = func(cmd *cobra.Command, args []string) error {
			calledLegacy = true
			return nil
		}
		parityDiffRunner = func(cmd *cobra.Command, args []string) error {
			calledParity = true
			return nil
		}

		if err := runDiff(cmd, nil); err != nil {
			t.Fatalf("runDiff: %v", err)
		}
		if !calledLegacy || calledParity {
			t.Fatalf("expected legacy path only, got legacy=%v parity=%v", calledLegacy, calledParity)
		}
	})

	t.Run("parity opt-in", func(t *testing.T) {
		cmd := diffCmd()
		if err := cmd.Flags().Set(diffModeFlag, diffModeParity); err != nil {
			t.Fatalf("set mode flag: %v", err)
		}
		if err := cmd.Flags().Set("left", "left.c1z"); err != nil {
			t.Fatalf("set left flag: %v", err)
		}
		if err := cmd.Flags().Set("right", "right.c1z"); err != nil {
			t.Fatalf("set right flag: %v", err)
		}

		calledLegacy := false
		calledParity := false
		legacyDiffRunner = func(cmd *cobra.Command, args []string) error {
			calledLegacy = true
			return nil
		}
		parityDiffRunner = func(cmd *cobra.Command, args []string) error {
			calledParity = true
			return nil
		}

		if err := runDiff(cmd, nil); err != nil {
			t.Fatalf("runDiff: %v", err)
		}
		if calledLegacy || !calledParity {
			t.Fatalf("expected parity path only, got legacy=%v parity=%v", calledLegacy, calledParity)
		}
	})

	t.Run("left right do not disable legacy by themselves", func(t *testing.T) {
		cmd := diffCmd()
		if err := cmd.Flags().Set("left", "left.c1z"); err != nil {
			t.Fatalf("set left flag: %v", err)
		}
		if err := cmd.Flags().Set("right", "right.c1z"); err != nil {
			t.Fatalf("set right flag: %v", err)
		}

		err := runDiff(cmd, nil)
		if err == nil {
			t.Fatal("expected error when left/right are provided without parity mode")
		}
		if !strings.Contains(err.Error(), "--left/--right require --mode parity") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("parity mode requires left and right", func(t *testing.T) {
		cmd := diffCmd()
		if err := cmd.Flags().Set(diffModeFlag, diffModeParity); err != nil {
			t.Fatalf("set mode flag: %v", err)
		}

		err := runDiff(cmd, nil)
		if err == nil {
			t.Fatal("expected error when parity mode is missing left/right")
		}
		if !strings.Contains(err.Error(), "both --left and --right are required with --mode parity") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunParityDiffPassesFiltersAndUsesGroupedOutput(t *testing.T) {
	originalLoader := loadParitySnapshotFn
	originalRenderer := renderParityDiffOutputFn
	defer func() {
		loadParitySnapshotFn = originalLoader
		renderParityDiffOutputFn = originalRenderer
	}()

	var loadedPaths []string
	var loadedOptions []parityOptions
	loadParitySnapshotFn = func(ctx context.Context, c1zPath string, options parityOptions) (*paritySnapshot, error) {
		loadedPaths = append(loadedPaths, c1zPath)
		loadedOptions = append(loadedOptions, options)
		return &paritySnapshot{
			Resources:       map[string]*parityResourceRecord{},
			Entitlements:    map[string]*parityEntitlementRecord{},
			Grants:          map[string]*parityGrantRecord{},
			ResourceRefs:    map[string]struct{}{},
			EntitlementRefs: map[string]struct{}{},
		}, nil
	}

	renderCalled := false
	renderParityDiffOutputFn = func(outputFormat string, out *parityDiffOutput) error {
		renderCalled = true
		if outputFormat != "json" {
			t.Fatalf("expected json output format, got %q", outputFormat)
		}
		if out == nil {
			t.Fatal("expected parity output")
		}
		return nil
	}

	cmd := diffCmd()
	cmd.Flags().String("output-format", "console", "")
	mustSetFlag(t, cmd, diffModeFlag, diffModeParity)
	mustSetFlag(t, cmd, "left", "left.c1z")
	mustSetFlag(t, cmd, "right", "right.c1z")
	mustSetFlag(t, cmd, "scope", "grant")
	mustSetFlag(t, cmd, resourceTypeFlag, "enterprise_application,directory_role")
	mustSetFlag(t, cmd, "output-format", "json")

	if err := runParityDiff(cmd, nil); err != nil {
		t.Fatalf("runParityDiff: %v", err)
	}

	if !renderCalled {
		t.Fatal("expected renderer to be called")
	}
	if len(loadedPaths) != 2 || loadedPaths[0] != "left.c1z" || loadedPaths[1] != "right.c1z" {
		t.Fatalf("unexpected loaded paths: %#v", loadedPaths)
	}
	if len(loadedOptions) != 2 {
		t.Fatalf("expected two load calls, got %d", len(loadedOptions))
	}
	for _, options := range loadedOptions {
		if options.Scope != parityScopeGrant {
			t.Fatalf("expected grant scope, got %q", options.Scope)
		}
		if !options.matchesResourceType("enterprise_application") || !options.matchesResourceType("directory_role") || options.matchesResourceType("user") {
			t.Fatalf("unexpected resource type filter: %#v", options.ResourceTypes)
		}
	}
}

func TestRenderParityConsoleGroupsOutputForTriage(t *testing.T) {
	out := buildRepresentativeParityOutput(t)

	consoleOutput := captureStdout(t, func() {
		if err := renderParityConsole(out); err != nil {
			t.Fatalf("renderParityConsole: %v", err)
		}
	})

	expectedSnippets := []string{
		"Summary",
		"Likely Causes",
		"Enterprise Application Focus",
		"Missing Resources",
		"Missing Entitlements",
		"Missing Grants",
		"Field Mismatches",
		"Dangling References",
		"enterprise_application: left=0 right=3",
		"principal type mismatches:",
		"managed_identity -> service_principal: 2",
	}
	for _, snippet := range expectedSnippets {
		if !strings.Contains(consoleOutput, snippet) {
			t.Fatalf("expected console output to contain %q, got:\n%s", snippet, consoleOutput)
		}
	}
}

func TestMarshalParityJSONIncludesGroupedShape(t *testing.T) {
	out := buildRepresentativeParityOutput(t)

	payload, err := marshalParityJSON(out)
	if err != nil {
		t.Fatalf("marshalParityJSON: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal parity json: %v", err)
	}

	for _, key := range []string{"summary", "groupedByFamily", "groupedByResourceType", "likelyCauses", "resources", "entitlements", "grants"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("expected json key %q in payload: %s", key, string(payload))
		}
	}

	families := decoded["groupedByFamily"].(map[string]any)
	if _, ok := families["missingResources"]; !ok {
		t.Fatalf("expected grouped missingResources: %s", string(payload))
	}

	grants := decoded["grants"].(map[string]any)
	dangling := grants["danglingReferences"].(map[string]any)
	right := dangling["right"].([]any)
	firstRight := right[0].(map[string]any)
	if firstRight["reason"] == nil {
		t.Fatalf("expected typed dangling reason in json: %s", string(payload))
	}
}

func TestRepresentativeRootCauseGrouping(t *testing.T) {
	out := buildRepresentativeParityOutput(t)

	if len(out.LikelyCauses) == 0 {
		t.Fatal("expected likely causes")
	}

	var sawInventory bool
	var sawPrincipalType bool
	var sawDanglingEntitlement bool
	for _, cause := range out.LikelyCauses {
		switch cause.Category {
		case "inventory_gap":
			sawInventory = true
		case "principal_type_mismatch":
			sawPrincipalType = true
		case "dangling_entitlement_refs":
			sawDanglingEntitlement = true
		}
	}

	if !sawInventory || !sawPrincipalType || !sawDanglingEntitlement {
		t.Fatalf("expected inventory/principal-type/dangling-entitlement causes, got %#v", out.LikelyCauses)
	}
}

func TestDiffNormalizedValuesDistinguishesMissingEmptyAndType(t *testing.T) {
	missingVsEmpty := diffNormalizedValues("field", parityMissingValue{}, "")
	if len(missingVsEmpty) != 1 || missingVsEmpty[0].Left != "<missing>" || missingVsEmpty[0].Right != "" {
		t.Fatalf("expected missing vs empty difference, got %#v", missingVsEmpty)
	}

	typeMismatch := diffNormalizedValues("field", "true", true)
	if len(typeMismatch) != 1 {
		t.Fatalf("expected type difference, got %#v", typeMismatch)
	}
}

func TestNormalizationCoversAdditionalAnnotationFamilies(t *testing.T) {
	resourceRecord, err := newParityResourceRecord(&v2.Resource{
		Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"},
		Annotations: mustAnnotations(t,
			&v2.GroupTrait{Profile: mustStruct(t, map[string]any{"group": "ops"})},
			&v2.RoleTrait{Profile: mustStruct(t, map[string]any{"scope": "global"})},
			&v2.ScopeBindingTrait{
				RoleId:          &v2.ResourceId{ResourceType: "role", Resource: "role-1"},
				ScopeResourceId: &v2.ResourceId{ResourceType: "project", Resource: "proj-1"},
			},
			&v2.SecretTrait{
				Profile:     mustStruct(t, map[string]any{"secret": "token"}),
				CreatedAt:   timestamppb.Now(),
				CreatedById: &v2.ResourceId{ResourceType: "user", Resource: "user-1"},
			},
		),
	})
	if err != nil {
		t.Fatalf("newParityResourceRecord: %v", err)
	}

	entitlementRecord, err := newParityEntitlementRecord(&v2.Entitlement{
		Id:       "ent-1",
		Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"}},
		Annotations: mustAnnotations(t, &v2.EntitlementImmutable{
			SourceId: "source-entitlement",
			Metadata: mustStruct(t, map[string]any{"immutable": true}),
		}),
	})
	if err != nil {
		t.Fatalf("newParityEntitlementRecord: %v", err)
	}

	grantRecord, err := newParityGrantRecord(&v2.Grant{
		Id: "grant-1",
		Entitlement: &v2.Entitlement{
			Id:       "ent-1",
			Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"}},
		},
		Principal: &v2.Resource{Id: &v2.ResourceId{ResourceType: "user", Resource: "user-1"}},
		Annotations: mustAnnotations(t, &v2.GrantImmutable{
			SourceId: "source-grant",
			Metadata: mustStruct(t, map[string]any{"immutable": true}),
		}),
	})
	if err != nil {
		t.Fatalf("newParityGrantRecord: %v", err)
	}

	for _, key := range []string{"groupTrait", "roleTrait", "scopeBindingTrait", "secretTrait"} {
		if _, ok := resourceRecord.Annotations[key]; !ok {
			t.Fatalf("expected resource annotation %q in %#v", key, resourceRecord.Annotations)
		}
	}
	if _, ok := entitlementRecord.Annotations["entitlementImmutable"]; !ok {
		t.Fatalf("expected entitlementImmutable annotation in %#v", entitlementRecord.Annotations)
	}
	if _, ok := grantRecord.Annotations["grantImmutable"]; !ok {
		t.Fatalf("expected grantImmutable annotation in %#v", grantRecord.Annotations)
	}
}

func TestNormalizationPreservesUnknownAnnotations(t *testing.T) {
	record, err := newParityGrantRecord(&v2.Grant{
		Id: "grant-1",
		Entitlement: &v2.Entitlement{
			Id:       "ent-1",
			Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"}},
		},
		Principal: &v2.Resource{Id: &v2.ResourceId{ResourceType: "user", Resource: "user-1"}},
		Annotations: mustAnnotations(t,
			&v2.GrantAlreadyExists{},
		),
	})
	if err != nil {
		t.Fatalf("newParityGrantRecord: %v", err)
	}

	if _, ok := record.Annotations["grantAlreadyExists"]; !ok {
		t.Fatalf("expected unknown annotation to be preserved, got %#v", record.Annotations)
	}
}

func TestNormalizationDoesNotSortArbitraryRepeatedFields(t *testing.T) {
	left, err := normalizeComparableValue(map[string]any{
		"ordered": []any{"a", "b"},
	})
	if err != nil {
		t.Fatalf("normalizeComparableValue left: %v", err)
	}
	right, err := normalizeComparableValue(map[string]any{
		"ordered": []any{"b", "a"},
	})
	if err != nil {
		t.Fatalf("normalizeComparableValue right: %v", err)
	}

	differences := diffNormalizedValues("", left, right)
	if len(differences) == 0 {
		t.Fatal("expected ordered repeated field difference")
	}

	leftUnordered, err := normalizeComparableValue(map[string]any{
		"grantableTo": []any{"a", "b"},
	})
	if err != nil {
		t.Fatalf("normalizeComparableValue left unordered: %v", err)
	}
	rightUnordered, err := normalizeComparableValue(map[string]any{
		"grantableTo": []any{"b", "a"},
	})
	if err != nil {
		t.Fatalf("normalizeComparableValue right unordered: %v", err)
	}
	if differences := diffNormalizedValues("", leftUnordered, rightUnordered); len(differences) != 0 {
		t.Fatalf("expected grantableTo ordering to be ignored, got %#v", differences)
	}
}

func TestValidateGrantReferencesDetectsMissingGrantResource(t *testing.T) {
	grant := &parityGrantRecord{
		Key: "id:grant-2",
		Tuple: parityGrantTuple{
			Resource:      &parityResourceKey{ResourceTypeID: "enterprise_application", ResourceID: "app-2"},
			EntitlementID: "ent-2",
			Principal:     &parityResourceKey{ResourceTypeID: "user", ResourceID: "user-2"},
		},
	}

	dangling := validateGrantReferences(grant, &paritySnapshot{
		ResourceRefs:    map[string]struct{}{"user:user-2": {}},
		EntitlementRefs: map[string]struct{}{"ent-2": {}},
	})

	if len(dangling) != 1 || dangling[0].Reason != "missing_grant_resource" {
		t.Fatalf("expected missing grant resource dangling ref, got %#v", dangling)
	}
}

func TestCompareGrantRecordSetsKeepsExtraDuplicateTuple(t *testing.T) {
	baseGrant, err := newParityGrantRecord(&v2.Grant{
		Id: "grant-a",
		Entitlement: &v2.Entitlement{
			Id:       "ent-1",
			Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: "enterprise_application", Resource: "app-1"}},
		},
		Principal: &v2.Resource{Id: &v2.ResourceId{ResourceType: "user", Resource: "user-1"}},
	})
	if err != nil {
		t.Fatalf("newParityGrantRecord: %v", err)
	}
	duplicateGrant := *baseGrant
	duplicateGrant.Key = "id:grant-b"
	duplicateGrant.ID = "grant-b"

	out, err := compareGrantRecordSets(
		map[string]*parityGrantRecord{baseGrant.Key: baseGrant, duplicateGrant.Key: &duplicateGrant},
		map[string]*parityGrantRecord{baseGrant.Key: baseGrant},
	)
	if err != nil {
		t.Fatalf("compareGrantRecordSets: %v", err)
	}

	if len(out.OnlyInLeft) != 1 || out.OnlyInLeft[0].Key != "id:grant-b" {
		t.Fatalf("expected duplicate tuple to remain only in left, got %#v", out.OnlyInLeft)
	}
}

func TestDetectPrincipalTypeMismatchesRequiresUnambiguousTypes(t *testing.T) {
	leftA := &parityGrantRecord{
		Key: "left-a",
		Tuple: parityGrantTuple{
			Resource:      &parityResourceKey{ResourceTypeID: "enterprise_application", ResourceID: "app-1"},
			EntitlementID: "ent-1",
			Principal:     &parityResourceKey{ResourceTypeID: "managed_identity", ResourceID: "principal-1"},
		},
	}
	leftB := &parityGrantRecord{
		Key: "left-b",
		Tuple: parityGrantTuple{
			Resource:      &parityResourceKey{ResourceTypeID: "enterprise_application", ResourceID: "app-1"},
			EntitlementID: "ent-1",
			Principal:     &parityResourceKey{ResourceTypeID: "user", ResourceID: "principal-1"},
		},
	}
	rightA := &parityGrantRecord{
		Key: "right-a",
		Tuple: parityGrantTuple{
			Resource:      &parityResourceKey{ResourceTypeID: "enterprise_application", ResourceID: "app-1"},
			EntitlementID: "ent-1",
			Principal:     &parityResourceKey{ResourceTypeID: "service_principal", ResourceID: "principal-1"},
		},
	}

	mismatches := detectPrincipalTypeMismatches([]*parityGrantRecord{leftA, leftB}, []*parityGrantRecord{rightA})
	if len(mismatches) != 0 {
		t.Fatalf("expected ambiguous buckets to be skipped, got %#v", mismatches)
	}
}

func containsField(differences []parityFieldDifference, field string) bool {
	for _, difference := range differences {
		if difference.Field == field {
			return true
		}
	}
	return false
}

func mustAnnotations(t *testing.T, messages ...proto.Message) []*anypb.Any {
	t.Helper()

	ret := make([]*anypb.Any, 0, len(messages))
	for _, message := range messages {
		annotation, err := anypb.New(message)
		if err != nil {
			t.Fatalf("build annotation: %v", err)
		}
		ret = append(ret, annotation)
	}
	return ret
}

func buildRepresentativeParityOutput(t *testing.T) *parityDiffOutput {
	t.Helper()

	makeResource := func(resourceTypeID, resourceID, name string) *parityResourceRecord {
		t.Helper()
		record, err := newParityResourceRecord(&v2.Resource{
			Id:          &v2.ResourceId{ResourceType: resourceTypeID, Resource: resourceID},
			DisplayName: name,
		})
		if err != nil {
			t.Fatalf("newParityResourceRecord: %v", err)
		}
		return record
	}

	makeEntitlement := func(resourceTypeID, resourceID, entitlementID, name string) *parityEntitlementRecord {
		t.Helper()
		record, err := newParityEntitlementRecord(&v2.Entitlement{
			Id:          entitlementID,
			DisplayName: name,
			Resource:    &v2.Resource{Id: &v2.ResourceId{ResourceType: resourceTypeID, Resource: resourceID}},
		})
		if err != nil {
			t.Fatalf("newParityEntitlementRecord: %v", err)
		}
		return record
	}

	makeGrant := func(id, resourceTypeID, resourceID, entitlementID, principalTypeID, principalID string) *parityGrantRecord {
		t.Helper()
		record, err := newParityGrantRecord(&v2.Grant{
			Id: id,
			Entitlement: &v2.Entitlement{
				Id:       entitlementID,
				Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: resourceTypeID, Resource: resourceID}},
			},
			Principal: &v2.Resource{Id: &v2.ResourceId{ResourceType: principalTypeID, Resource: principalID}},
		})
		if err != nil {
			t.Fatalf("newParityGrantRecord: %v", err)
		}
		return record
	}

	leftGrant1 := makeGrant("left-grant-1", "enterprise_application", "app-a", "ent-shared", "managed_identity", "principal-1")
	leftGrant2 := makeGrant("left-grant-2", "enterprise_application", "app-a", "ent-shared", "managed_identity", "principal-2")
	leftGrant3 := makeGrant("left-grant-3", "enterprise_application", "app-a", "ent-left", "user", "principal-3")
	rightEntitlement1 := makeEntitlement("enterprise_application", "app-a", "ent-right-1", "Right Entitlement 1")
	rightEntitlement2 := makeEntitlement("enterprise_application", "app-a", "ent-right-2", "Right Entitlement 2")
	rightGrant1 := makeGrant("right-grant-1", "enterprise_application", "app-a", "ent-shared", "service_principal", "principal-1")
	rightGrant2 := makeGrant("right-grant-2", "enterprise_application", "app-a", "ent-shared", "service_principal", "principal-2")
	rightGrant3 := makeGrant("right-grant-3", "enterprise_application", "app-a", "ent-right-1", "user", "principal-4")

	leftSnapshot := &paritySnapshot{
		Resources: map[string]*parityResourceRecord{
			makeResource("enterprise_application", "app-a", "App A").sortKey(): makeResource("enterprise_application", "app-a", "App A"),
		},
		Entitlements: map[string]*parityEntitlementRecord{
			makeEntitlement("enterprise_application", "app-a", "ent-left", "Left Entitlement").sortKey(): makeEntitlement("enterprise_application", "app-a", "ent-left", "Left Entitlement"),
			makeEntitlement("enterprise_application", "app-a", "ent-shared", "Shared Left").sortKey():    makeEntitlement("enterprise_application", "app-a", "ent-shared", "Shared Left"),
		},
		Grants: map[string]*parityGrantRecord{
			leftGrant1.sortKey(): leftGrant1,
			leftGrant2.sortKey(): leftGrant2,
			leftGrant3.sortKey(): leftGrant3,
		},
		ResourceRefs: map[string]struct{}{
			"enterprise_application:app-a": {},
			"managed_identity:principal-1": {},
			"managed_identity:principal-2": {},
			"user:principal-3":             {},
		},
		EntitlementRefs: map[string]struct{}{
			"ent-left":   {},
			"ent-shared": {},
		},
	}

	rightSnapshot := &paritySnapshot{
		Resources: map[string]*parityResourceRecord{
			makeResource("enterprise_application", "app-a", "App A Renamed").sortKey(): makeResource("enterprise_application", "app-a", "App A Renamed"),
			makeResource("enterprise_application", "app-b", "App B").sortKey():         makeResource("enterprise_application", "app-b", "App B"),
			makeResource("enterprise_application", "app-c", "App C").sortKey():         makeResource("enterprise_application", "app-c", "App C"),
			makeResource("enterprise_application", "app-d", "App D").sortKey():         makeResource("enterprise_application", "app-d", "App D"),
		},
		Entitlements: map[string]*parityEntitlementRecord{
			rightEntitlement1.sortKey(): rightEntitlement1,
			rightEntitlement2.sortKey(): rightEntitlement2,
			makeEntitlement("enterprise_application", "app-a", "ent-shared", "Shared Right").sortKey(): makeEntitlement(
				"enterprise_application",
				"app-a",
				"ent-shared",
				"Shared Right",
			),
		},
		Grants: map[string]*parityGrantRecord{
			rightGrant1.sortKey(): rightGrant1,
			rightGrant2.sortKey(): rightGrant2,
			rightGrant3.sortKey(): rightGrant3,
		},
		ResourceRefs: map[string]struct{}{
			"enterprise_application:app-a":  {},
			"enterprise_application:app-b":  {},
			"enterprise_application:app-c":  {},
			"enterprise_application:app-d":  {},
			"service_principal:principal-1": {},
			"service_principal:principal-2": {},
			"user:principal-4":              {},
		},
		EntitlementRefs: map[string]struct{}{
			"ent-right-1": {},
			"ent-right-2": {},
			"ent-shared":  {},
		},
		DanglingGrants: []parityDanglingReference{
			{
				GrantKey:        "id:right-dangling-1",
				ResourceTypeID:  "enterprise_application",
				PrincipalTypeID: "user",
				Tuple: parityGrantTuple{
					Resource:      &parityResourceKey{ResourceTypeID: "enterprise_application", ResourceID: "app-a"},
					EntitlementID: "missing-ent-1",
					Principal:     &parityResourceKey{ResourceTypeID: "user", ResourceID: "principal-5"},
				},
				Reason:    "missing_entitlement",
				Reference: "entitlement",
				Value:     "missing-ent-1",
			},
			{
				GrantKey:        "id:right-dangling-2",
				ResourceTypeID:  "enterprise_application",
				PrincipalTypeID: "user",
				Tuple: parityGrantTuple{
					Resource:      &parityResourceKey{ResourceTypeID: "enterprise_application", ResourceID: "app-a"},
					EntitlementID: "missing-ent-2",
					Principal:     &parityResourceKey{ResourceTypeID: "user", ResourceID: "principal-6"},
				},
				Reason:    "missing_entitlement",
				Reference: "entitlement",
				Value:     "missing-ent-2",
			},
		},
	}

	out, err := compareParitySnapshots(leftSnapshot, rightSnapshot)
	if err != nil {
		t.Fatalf("compareParitySnapshots: %v", err)
	}
	return out
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll: %v", err)
	}
	return string(output)
}

func mustSetFlag(t *testing.T, cmd *cobra.Command, name string, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set %s flag: %v", name, err)
	}
}

func mustStruct(t *testing.T, value map[string]any) *structpb.Struct {
	t.Helper()

	ret, err := structpb.NewStruct(value)
	if err != nil {
		t.Fatalf("build struct: %v", err)
	}
	return ret
}
