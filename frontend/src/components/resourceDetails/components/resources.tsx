import React, { Fragment } from "react";
import { ListItem } from "..";

export const ResourceDetails = ({ resource }) => (
    <Fragment>
      {resource.annotations?.map((annotation) => {
        return (
          annotation?.profile &&
          Object.keys(annotation?.profile).map((key) => (
            <ListItem key={key} label={key} value={annotation.profile[key]} />
          ))
        );
      })}

      <ListItem label="Id" value={resource.id.resource} />
      <ListItem label="Resource type" value={resource.id.resource_type} />
    </Fragment>
);
