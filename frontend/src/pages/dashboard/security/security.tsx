import React from "react";
import { LayoutWrapper } from "../components/styles";
import { SecurityScoreCard } from "./securityScoreCard";
import { DirectAssignments } from "./directAssignments/directAssignments";
import { CardWithGraph } from "./cardWithGraph/cardWithGraph";
import { styled } from "@mui/material";

// Placeholder data.
const inactiveUsers = [
  {
    resource: { id: "12345", display_name: "Test User" },
    resource_type: { id: "user" },
  },
  {
    resource: { id: "54321", display_name: "Test User 2" },
    resource_type: { id: "user" },
  },
];

const inactiveUsersData = [
  { name: "users", value: 63 },
  { name: "inactive", value: 23 },
];

const nonMfaUsers = [
  {
    resource: { id: "121212", display_name: "Test User 3" },
    resource_type: { id: "user" },
  },
  {
    resource: { id: "534534", display_name: "Test User 4" },
    resource_type: { id: "user" },
  },
  {
    resource: { id: "32342", display_name: "Test User 5" },
    resource_type: { id: "user" },
  },
];

const nonMfaUsersData = [
  { name: "users", value: 63 },
  { name: "nonMfa", value: 22 },
];

const nonSsoUsers = [
  {
    resource: { id: "543535", display_name: "Test User 6" },
    resource_type: { id: "user" },
  },
  {
    resource: { id: "543535", display_name: "Test User 7" },
    resource_type: { id: "user" },
  },
];

const nonSsoUsersData = [
  { name: "users", value: 63 },
  { name: "nonMfa", value: 43 },
];

const Container = styled("div")(() => ({
  display: "flex",
  justifyContent: "center",
  flexWrap: "wrap",
  width: "100%",
  margin: "8px",

  "> div": {
    margin: "8px"
  }
}));


export const Security = () => {
  return (
    <LayoutWrapper>
      <Container>
        <SecurityScoreCard score={78} />
        <DirectAssignments />
      </Container>
      <Container>
      <CardWithGraph
        header="inactive accounts"
        count={32}
        users={inactiveUsers}
        chartData={inactiveUsersData}
        text="21%"
      />
      <CardWithGraph
        header="non-mfa enabled accounts"
        count={12}
        users={nonMfaUsers}
        chartData={nonMfaUsersData}
        text="34%"
      />
      <CardWithGraph
        header="non-sso enabled accounts"
        count={24}
        users={nonSsoUsers}
        chartData={nonSsoUsersData}
        text="12%"
      />
      </Container>
    </LayoutWrapper>
  );
};
