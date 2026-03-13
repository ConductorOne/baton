import React, { Fragment, useMemo } from "react";
import { Close, LinkIcon } from "../icons/icons";
import { Button, Typography } from "@mui/material";
import {
  Container,
  Details,
  Label,
  ResourceDetailsPanel,
  StyledDiv,
  Value,
  CloseButton,
  ModalHeader,
  StyledLink,
} from "./styles";
import { normalizeString } from "../../common/helpers";
import { EntitlementDetails } from "./components/entitlements";
import { ResourceDetails } from "./components/resources";

export const ListItem = ({ label, value }) => {
  return (
    <Container>
      {value && (
        <Fragment>
          <Label>{normalizeString(label, true)}</Label>
          <Value>{value}</Value>
        </Fragment>
      )}
    </Container>
  );
};

export const ResourceDetailsModal = ({
  resource,
  resourceDetails,
  closeDetails,
  anchorPosition,
}) => {
  const panelStyle = useMemo(() => {
    if (!anchorPosition) {
      // Fallback: top-right corner
      return { top: 80, right: 16 };
    }
    const panelWidth = 336;
    const panelMaxHeight = window.innerHeight * 0.7;
    let left = anchorPosition.left + 16;
    let top = anchorPosition.top - 20;

    // Flip to left of click if it would overflow right edge
    if (left + panelWidth > window.innerWidth - 16) {
      left = anchorPosition.left - panelWidth - 16;
    }
    // Clamp to not overflow bottom
    if (top + panelMaxHeight > window.innerHeight - 16) {
      top = window.innerHeight - panelMaxHeight - 16;
    }
    // Clamp to not go above viewport
    if (top < 16) top = 16;

    return { top, left };
  }, [anchorPosition]);

  return (
    <ResourceDetailsPanel style={panelStyle}>
      <ModalHeader>
        <StyledDiv>
          <Typography variant="h5">
            {resource.display_name || resource.resource?.display_name}
          </Typography>
          <CloseButton onClick={closeDetails}>
            <Close />
          </CloseButton>
        </StyledDiv>
        {resourceDetails.resourceOpened && (
          <Button
            variant="text"
            startIcon={<LinkIcon />}
            disableElevation
          >
            <StyledLink
              to={`/${resource.resource_type.id}/${resource.resource.id.resource}`}
            >
              focus
            </StyledLink>
          </Button>
        )}
      </ModalHeader>
      <Details>
        {resourceDetails.entitlementOpened && (
          <EntitlementDetails entitlement={resource} />
        )}
        {resourceDetails.resourceOpened && (
          <ResourceDetails resource={resource.resource} profile={resource.profile} />
        )}
      </Details>
    </ResourceDetailsPanel>
  );
};
