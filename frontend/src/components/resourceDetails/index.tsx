import React, { Fragment } from "react";
import { Close, Link } from "../icons/icons";
import { useTheme } from "@mui/material/styles";
import { Button, Typography } from "@mui/material";
import {
  Container,
  Details,
  Label,
  ResourceDetailsDrawer,
  StyledDiv,
  Value,
  CloseButton,
  ModalHeader,
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

export const ResourceDetailsModal = ({ resource, resourceDetails, closeDetails }) => {
  const theme = useTheme();
  const copyToClipboard = () => {
    navigator.clipboard.writeText(`http://localhost:3000/${resource.resource_type.id}/${resource.resource.id.resource}`)
  }

  return (
    <ResourceDetailsDrawer theme={theme} variant="permanent" anchor="right">
      <ModalHeader>
        <StyledDiv>
          <Typography variant="h5">
            {resource.display_name || resource.resource?.display_name}
          </Typography>
          <CloseButton onClick={closeDetails}>
            <Close color="secondary" />
          </CloseButton>
        </StyledDiv>
        {resourceDetails.resourceOpened && (
          <Button
            variant="text"
            color="secondary"
            startIcon={<Link color="secondary" />}
            disableElevation
            onClick={copyToClipboard}
          >
            Copy Link
          </Button>
        )}
      </ModalHeader>
      <Details>
        {resourceDetails.entitlementOpened && (
          <EntitlementDetails entitlement={resource} />
        )}
        { resourceDetails.resourceOpened && (
          <ResourceDetails resource={resource.resource} />
        )}
      </Details>
    </ResourceDetailsDrawer>
  );
};
