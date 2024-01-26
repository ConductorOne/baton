import React, { Fragment } from "react";
import { ListItem } from "..";
import { Label, Value } from "../styles";

const purposeMap = {
  0: "Unspecified",
  1: "Assignment",
  2: "Permission",
};

export const EntitlementDetails = ({
  entitlement,
}) => {
  return (
    <Fragment>
      {entitlement.description && (
        <ListItem
          key={entitlement.description}
          label="Description"
          value={entitlement.description}
        />
      )}

      {entitlement.purpose && (
        <ListItem
          key={entitlement.purpose}
          label="Purpose"
          value={purposeMap[entitlement.purpose]}
        />
      )}

      {entitlement.slug && (
        <ListItem
          key={entitlement.slug}
          label="slug"
          value={entitlement.slug}
        />
      )}

      {entitlement.id && (
        <ListItem
          key={entitlement.id}
          label="id"
          value={entitlement.id}
        />
      )}

      {entitlement?.grantable_to && (
        <Fragment>
          <Label>Grantable to</Label>
          {entitlement.grantable_to.map((resource) => (
            <Value key={resource.display_name}>{resource.display_name}</Value>
          ))}
        </Fragment>
      )}
    </Fragment>
  );
}
