import React from "react";
import { LayoutWrapper } from "../components/styles";
import { SecurityScoreCard } from "./securityScoreCard";
import { DirectAssignments } from "./directAssignments/directAssignments";
import { CardWithGraph } from "./cardWithGraph/cardWithGraph";
import { styled } from "@mui/material";

const inactiveUsers = [
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

const inactiveUsersData = [
  { name: "users", value: 63 },
  { name: "inactive", value: 23 },
];

const nonMfaUsers = [
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

const nonMfaUsersData = [
  { name: "users", value: 63 },
  { name: "nonMfa", value: 22 },
];


const nonSsoUsers = [
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

const nonSsoUsersData = [
  { name: "users", value: 63 },
  { name: "nonMfa", value: 43 },
];

const Container = styled("div")(() => ({
  display: "flex",
  justifyContent: "center",
  flexWrap: "wrap",

  "> div": {
    marginRight: "16px"
  }
}));


export const Security = () => {
  return (
    <LayoutWrapper>
      <Container>
        <SecurityScoreCard score={78} />
        <DirectAssignments />
      </Container>
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
    </LayoutWrapper>
  );
};
