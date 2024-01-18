import React, { Fragment } from "react";
import { Back } from "../../icons/icons";
import { Divider, Typography, useTheme } from "@mui/material";
import pluralize from "pluralize";
import { normalizeString } from "../../../common/helpers";
import {
  ResourceLabel,
  ResourcesListWrapper,
  Sidebar,
  SidebarHeader,
  StyledButton,
  EmptyResourceLabel,
} from "../styles/styles";
import { useResources } from "../../../context/resources";

export const ResourcesSidebar = ({
  closeResourceList,
  resourceType,
  openTreeView,
}) => {
  const resources = useResources();
  const theme = useTheme();
  const [selectedResourceId, setSelectedResourceId] = React.useState(null);

  const handleListItemClick = async (resource) => {
    setSelectedResourceId(resource.resource.id.resource);
    await openTreeView(resource);
  };

  return (
    <Fragment>
      {resourceType && (
        <Sidebar theme={theme} variant="permanent">
          <SidebarHeader>
            <Typography variant="h5" color="primary.contrastText">
              {pluralize(normalizeString(resourceType, true))}
            </Typography>
            <StyledButton onClick={closeResourceList}>
              <Back />
            </StyledButton>
          </SidebarHeader>
          <Divider />
          <ResourcesListWrapper>
            {resources.mappedResources[resourceType] ? (
              resources.mappedResources[resourceType].map(resource => (
                <ResourceLabel
                  disableGutters
                  key={resource.resource.id.resource}
                  selected={selectedResourceId === resource.resource.id.resource}
                  onClick={async () => await handleListItemClick(resource)}
                >
                  <Typography color="primary.contrastText">
                    {resource.resource.display_name}
                  </Typography>
                </ResourceLabel>
              ))
            ) : (
              <EmptyResourceLabel color="primary.contrastText">
                No resources
              </EmptyResourceLabel>
            )}
          </ResourcesListWrapper>
        </Sidebar>
      )}
    </Fragment>
  );
};
