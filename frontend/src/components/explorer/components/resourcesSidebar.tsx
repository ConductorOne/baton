import React, { Fragment } from "react";
import { Back } from "../../icons/icons";
import { Typography, useTheme } from "@mui/material";
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
  const [selectedIndex, setSelectedIndex] = React.useState(null);

  const handleListItemClick = async (index: number, resource) => {
    setSelectedIndex(index);
    await openTreeView(resource);
  };

  return (
    <Fragment>
      {resourceType && (
        <Sidebar theme={theme} variant="permanent">
          <SidebarHeader>
            <Typography variant="h6" color="primary.contrastText">
              {pluralize(normalizeString(resourceType, true))}
            </Typography>
            <StyledButton onClick={closeResourceList}>
              <Back color="icon" size="medium" />
            </StyledButton>
          </SidebarHeader>
          <ResourcesListWrapper>
            {resources.mappedResources[resourceType] ? (
              resources.mappedResources[resourceType].map((resource, i) => (
                <ResourceLabel
                  disableGutters
                  key={resource.resource.id.resource}
                  selected={selectedIndex === i}
                  onClick={async () => await handleListItemClick(i, resource)}
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
