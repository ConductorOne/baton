import React from "react";
import { Card } from "../components/cards/cards";
import { Resources } from "./resourcesGraph/resourcesGraph";
import { LayoutWrapper } from "../components/styles";
import { ResourcesListCard } from "./resourcesListCard/resourcesList";
import { IdentityWrapper } from "./styles";

const identityResources = [
  {
    name: "users",
    quantity: 200,
  },
  {
    name: "service accounts",
    quantity: 54,
  },
  {
    name: "api keys",
    quantity: 432,
  },
  {
    name: "system accounts",
    quantity: 20,
  },
];

const rolesMock = [
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

const groupsMock = [
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

export const Inventory = () => {
  return (
    <LayoutWrapper>
      <IdentityWrapper>
        <Card count="706" size="l" label="Identities" withMargin />
        {identityResources.map((resource) => (
          <Card
            withMargin
            isColumn
            key={resource.name}
            label={resource.name}
            count={resource.quantity.toString()}
            size="s"
          />
        ))}
      </IdentityWrapper>
      <ResourcesListCard
        resourcesCount={34}
        resourceType={"Roles"}
        resources={rolesMock}
      />
      <ResourcesListCard
        resourcesCount={3}
        resourceType={"Groups"}
        resources={groupsMock}
      />
      <Resources />
    </LayoutWrapper>
  );
};
