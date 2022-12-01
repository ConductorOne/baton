package main

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	v1 "github.com/conductorone/baton/pb/baton/v1"
	"github.com/conductorone/baton/pkg/output"
	"github.com/conductorone/baton/pkg/storecache"
	"github.com/spf13/cobra"
)

func principalsCompareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare principals between two entitlements. These entitlements can exist in different C1Z files. The results are relative to the 'base' entitlement.",
		RunE:  runPrincipalsCompare,
	}

	// Filter by entitlement
	addEntitlementFlag(cmd)
	cmd.Flags().String("compare-file", "sync.c1z", "The path to the c1z file to compare with")
	cmd.Flags().String("compare-entitlement", "", "The entitlement ID from the secondary C1Z to compare against")

	return cmd
}

func runPrincipalsCompare(cmd *cobra.Command, args []string) error {
	ctx, err := logging.Init(context.Background(), "console", "error")
	if err != nil {
		return err
	}
	c1zPath, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	outputFormat, err := cmd.Flags().GetString("output-format")
	if err != nil {
		return err
	}
	outputManager := output.NewManager(ctx, outputFormat)

	m, err := manager.New(ctx, c1zPath)
	if err != nil {
		return err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return err
	}

	sc := storecache.NewStoreCache(ctx, store)

	compareC1zPath, err := cmd.Flags().GetString("compare-file")
	if err != nil {
		return err
	}

	mCompare, err := manager.New(ctx, compareC1zPath)
	if err != nil {
		return err
	}
	defer m.Close(ctx)

	storeCompare, err := mCompare.LoadC1Z(ctx)
	if err != nil {
		return err
	}

	scDiff := storecache.NewStoreCache(ctx, storeCompare)

	baseEntitlementID, err := cmd.Flags().GetString(entitlementFlag)
	if err != nil {
		return err
	}

	diffEntitlementID, err := cmd.Flags().GetString("compare-entitlement")
	if err != nil {
		return err
	}

	outputs := &v1.PrincipalsCompareOutput{}

	basePrincipals := make(map[string]*v1.ResourceOutput)
	pageToken := ""
	for {
		var principals []*v2.Resource
		principals, pageToken, err = listPrincipalsForEntitlement(ctx, baseEntitlementID, sc, pageToken)
		if err != nil {
			return err
		}

		for _, p := range principals {
			cacheKey := getResourceIdString(p)
			if _, ok := basePrincipals[cacheKey]; !ok {
				resourceType, err := sc.GetResourceType(ctx, p.Id.ResourceType)
				if err != nil {
					return err
				}

				var parent *v2.Resource
				if p.ParentResourceId != nil {
					parent, err = sc.GetResource(ctx, p.ParentResourceId)
					if err != nil {
						return err
					}
				}

				rOutput := &v1.ResourceOutput{
					Resource:     p,
					ResourceType: resourceType,
					Parent:       parent,
				}

				outputs.Base = append(outputs.Base, rOutput)
				basePrincipals[cacheKey] = rOutput
			}
		}

		if pageToken == "" {
			break
		}
	}

	diffPrincipals := make(map[string]*v1.ResourceOutput)
	pageToken = ""
	for {
		var principals []*v2.Resource
		principals, pageToken, err = listPrincipalsForEntitlement(ctx, diffEntitlementID, scDiff, pageToken)
		if err != nil {
			return err
		}

		for _, p := range principals {
			cacheKey := getResourceIdString(p)
			if _, ok := diffPrincipals[cacheKey]; !ok {
				resourceType, err := sc.GetResourceType(ctx, p.Id.ResourceType)
				if err != nil {
					return err
				}

				var parent *v2.Resource
				if p.ParentResourceId != nil {
					parent, err = sc.GetResource(ctx, p.ParentResourceId)
					if err != nil {
						return err
					}
				}

				rOutput := &v1.ResourceOutput{
					Resource:     p,
					ResourceType: resourceType,
					Parent:       parent,
				}

				outputs.Compared = append(outputs.Compared, rOutput)
				diffPrincipals[cacheKey] = rOutput
			}
		}

		if pageToken == "" {
			break
		}
	}

	matched := make(map[string]string)
	diffMatched := make(map[string]string)

	for baseID, basePrincipal := range basePrincipals {
		for diffID, diffPrincipal := range diffPrincipals {
			match, err := comparePrincipal(basePrincipal.Resource, diffPrincipal.Resource)
			if err != nil {
				return err
			}

			if match {
				matched[baseID] = diffID
				diffMatched[diffID] = baseID
				break
			}
		}

		if _, ok := matched[baseID]; !ok {
			outputs.Missing = append(outputs.Missing, basePrincipal)
		}
	}

	for pID, diffPrincipal := range diffPrincipals {
		if _, ok := diffMatched[pID]; !ok {
			outputs.Extra = append(outputs.Extra, diffPrincipal)
		}
	}

	err = outputManager.Output(ctx, outputs)
	if err != nil {
		return err
	}

	return nil
}

func comparePrincipal(base *v2.Resource, diff *v2.Resource) (bool, error) {
	// Return true if these are the same object
	if base.Id.ResourceType == diff.Id.ResourceType && base.Id.Resource == diff.Id.Resource {
		return true, nil
	}

	baseAnnos := annotations.Annotations(base.Annotations)
	diffAnnos := annotations.Annotations(diff.Annotations)

	baseUserTrait := &v2.UserTrait{}
	diffUserTrait := &v2.UserTrait{}

	baseHasUserTrait, err := baseAnnos.Pick(baseUserTrait)
	if err != nil {
		return false, err
	}

	diffHasUserTrait, err := diffAnnos.Pick(diffUserTrait)
	if err != nil {
		return false, err
	}

	switch {
	// Both resources have a user trait, try to compare email addresses
	case baseHasUserTrait && diffHasUserTrait:
		for _, baseEmail := range baseUserTrait.Emails {
			for _, diffEmail := range diffUserTrait.Emails {
				if baseEmail.Address == diffEmail.Address {
					return true, nil
				}
			}
		}

	// If they either don't both have a user trait or are both missing it, we know they aren't the same.
	case baseHasUserTrait != diffHasUserTrait:
		return false, nil
	}

	if base.DisplayName == diff.DisplayName {
		return true, nil
	}

	return false, nil
}
