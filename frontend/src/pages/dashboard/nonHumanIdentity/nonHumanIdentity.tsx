import React from "react";
import { Card } from "../components/cards/cards";
import { ResourcesListCard } from "../inventory/resourcesListCard/resourcesList";
import { styled } from "@mui/material";
import { colors } from "../../../style/colors";

const serviceAccounts = [
  {
    resource: { id: "bla", display_name: "Role 1" },
    resource_type: { id: "role" },
  },
  {
    resource: { id: "meh", display_name: "Role 2" },
    resource_type: { id: "role" },
  },
  {
    resource: { id: "wfg", display_name: "Role 3" },
    resource_type: { id: "role" },
  },
  {
    resource: { id: "wfg", display_name: "Role 4" },
    resource_type: { id: "role" },
  },
  {
    resource: { id: "wfg", display_name: "Role 5" },
    resource_type: { id: "role" },
  },
  {
    resource: { id: "wfg", display_name: "Role 6" },
    resource_type: { id: "role" },
  },
];

const systemAccounts = [
  {
    resource: { id: "sdasda", display_name: "group 1" },
    resource_type: { id: "role" },
  },
  {
    resource: { id: "asdasd", display_name: "group 2" },
    resource_type: { id: "role" },
  },
  {
    resource: { id: "wasdasdfg", display_name: "group 3" },
    resource_type: { id: "role" },
  },
];

export const Wrapper = styled("div", {
  shouldForwardProp: (prop) => prop !== "width",
})<{ width?: number}>(({ theme, width }) => ({
  display: "flex",
  flexDirection: "row",
  borderRadius: "16px",
  background: theme.palette.mode === "light" ? colors.white : colors.gray700,
  padding: "16px",
  width: "100%",
  maxWidth: "600px",
  height: "max-content",
  flexWrap: "wrap",

  "> div": {
    justifyContent: "space-evenly",
    width: 600,
  }
}));

const Container = styled("div")(() => ({
  display: "flex",
  flexDirection: "column",
}));

const Layout = styled("div")(() => ({
  display: "flex",
  flexDirection: "row",
  flexWrap: "wrap",
  padding: "25px",
  width: "100%",
  justifyContent: "center",

  "> div": {
    margin: "10px",
  },
}));

export const NonHumanIdentity = () => {
  return (
    <Layout>
      <ResourcesListCard
        resourceType="service accounts"
        resourcesCount={200}
        resources={serviceAccounts}
      />
      <ResourcesListCard
        resourceType="system accounts"
        resourcesCount={23}
        resources={systemAccounts}
      />
      <Container>
        <Wrapper width={600} style={{ marginBottom: "15px"}}>
          <Card count="1203" size="l" label="Oath Tokens" />
        </Wrapper>
        <Wrapper width={600}>
          <Card count="432" size="l" label="Api Tokens" />
        </Wrapper>
      </Container>
    </Layout>
  );
};
