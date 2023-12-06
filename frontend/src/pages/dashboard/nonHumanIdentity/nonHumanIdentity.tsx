import React, { useState } from "react";
import { Card } from "../components/cards/cards";
import { ResourcesListCard } from "../inventory/resourcesListCard/resourcesList";
import { styled } from "@mui/material";
import { colors } from "../../../style/colors";

export const Wrapper = styled("div", {
  shouldForwardProp: (prop) => prop !== "width",
})<{ width?: number }>(({ theme, width }) => ({
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
  },
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

  "> div": {
    margin: "10px",
  },
}));

export const NonHumanIdentity = () => {
  const [serviceAccounts, setServiceAccounts] = useState([]);
  const [systemAccounts, setSystemAccounts] = useState([]);
  return (
    <Layout>
      <ResourcesListCard
        resourceType="service accounts"
        resourcesCount={serviceAccounts?.length || 0}
        resources={serviceAccounts}
      />
      <ResourcesListCard
        resourceType="system accounts"
        resourcesCount={systemAccounts?.length || 0}
        resources={systemAccounts}
      />
      <Container>
        <Wrapper width={600} style={{ marginBottom: "15px" }}>
          <Card count="1203" size="l" label="Oath Tokens" />
        </Wrapper>
        <Wrapper width={600}>
          <Card count="432" size="l" label="Api Tokens" />
        </Wrapper>
      </Container>
    </Layout>
  );
};
