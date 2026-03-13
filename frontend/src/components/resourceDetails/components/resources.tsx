import React, { Fragment } from "react";
import { ListItem } from "..";

export const ResourceDetails = ({ resource, profile }) => (
    <Fragment>
      {profile && Object.keys(profile).map((key) => (
        <ListItem key={key} label={key} value={profile[key]} />
      ))}

      <ListItem label="Id" value={resource.id.resource} />
      <ListItem label="Resource type" value={resource.id.resource_type} />
    </Fragment>
);
