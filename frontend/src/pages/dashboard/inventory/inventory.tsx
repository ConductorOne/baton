import React, { useEffect, useState } from "react";
import { Card } from "../components/cards/cards";
import { Resources } from "./resourcesGraph/resourcesGraph";
import { LayoutWrapper } from "../components/styles";
import { ResourcesListCard } from "./resourcesListCard/resourcesList";
import { IdentityWrapper, LoadingWrapper, ResourcesWrapper } from "./styles";
import { useResources } from "../../../context/resources";
import { fetchResourcesWithUserCount } from "../../../components/explorer/api";
import { CircularProgress } from "@mui/material";
import { isObjectEmpty } from "../../../common/helpers";

export const Inventory = () => {
  const { identities, groupTraitTypes, roleTraitTypes } = useResources();
  const [roles, setRoles] = useState({});
  const [groups, setGroups] = useState({});
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      let roleData = {};
      for (const type of roleTraitTypes) {
        const response = await fetchResourcesWithUserCount(type);
        if (!roleData[type]) {
          roleData[type] = [];
        }
        roleData[type].push(...response.resources);
      }

      const groupData = {};
      for (const type of groupTraitTypes) {
        const response = await fetchResourcesWithUserCount(type);
        if (!groupData[type]) {
          groupData[type] = [];
        }
        groupData[type].push(...response.resources);
      }

      setRoles(roleData);
      setGroups(groupData);
      setIsLoading(false);
    };
    fetchData();
  }, [groupTraitTypes, roleTraitTypes]);

  return (
    <LayoutWrapper isColumn>
      {isLoading ? (
        <LoadingWrapper>
          <CircularProgress color="success" />
        </LoadingWrapper>
      ) : (
        <>
          <IdentityWrapper>
            <Card
              count={identities.count}
              size="l"
              label="Identities"
              withMargin
            />
            {identities.resourcesByType &&
              Object.keys(identities.resourcesByType).map((type) => (
                <Card
                  to={`/${type}`}
                  withMargin
                  isColumn
                  key={type}
                  label={type}
                  count={identities.resourcesByType[type]?.length}
                  size="s"
                />
              ))}
          </IdentityWrapper>
          <ResourcesWrapper>
            {!isObjectEmpty(roles) &&
              roleTraitTypes.map((type) => (
                <ResourcesListCard
                  resourcesCount={roles[type]?.length}
                  resourceType={type}
                  resources={roles[type]}
                />
              ))}
            {!isObjectEmpty(groups) &&
              groupTraitTypes.map((type) => (
                <ResourcesListCard
                  resourcesCount={groups[type]?.length}
                  resourceType={type}
                  resources={groups[type]}
                />
              ))}
            <Resources />
          </ResourcesWrapper>
        </>
      )}
    </LayoutWrapper>
  );
};
