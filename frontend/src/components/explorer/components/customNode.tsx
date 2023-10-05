import React from "react";
import { Handle, Position } from "reactflow";
import { Typography } from "@mui/material";
import { ResourceTypeIcon } from "../../icons/icons";
import { normalizeString } from "../../../common/helpers";
import { IconWrapper, NodeInfoWrapper, NodeWrapper } from "../styles/styles";

export const ChildNode = ({ data }) => (
  <>
    <Handle type="target" position={Position.Left} id={data.targetHandle} />
    <CustomNode data={data} />
  </>
);

export const ParentNode = ({ data }) => (
  <>
    <CustomNode data={data} />
    <Handle type="source" position={Position.Right} id={data.sourceHandle} />
  </>
);

export const CustomNode = ({ data }) => {
  return (
    <NodeWrapper>
      <IconWrapper className="iconWrapper">
        <ResourceTypeIcon
          resourceTrait={data.resourceTrait}
          color="icon"
          size="small"
        />
      </IconWrapper>
      <NodeInfoWrapper>
        <Typography>{data.label}</Typography>
        <Typography variant="caption">
          {normalizeString(data.resourceType, false)}
        </Typography>
      </NodeInfoWrapper>
    </NodeWrapper>
  );
};
