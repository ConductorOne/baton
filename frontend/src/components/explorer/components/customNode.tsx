import React from "react";
import { Handle, Position } from "reactflow";
import { Typography, useTheme } from "@mui/material";
import { normalizeString } from "../../../common/helpers";
import { IconWrapper, Node, NodeInfoWrapper, NodeWrapper } from "../styles/styles";
import { colors } from "../../../style/colors";
import { IconPerType, IconColors } from "../../icons/resourceTypeIcon";

export const ChildNode = ({ data, selected }) => {
  return (
    <Node isSelected={selected}>
      <Handle type="target" position={Position.Left} id={data.targetHandle} />
      <CustomNode data={data} />
    </Node>
  );
}

export const ExpandableGrantNode = ({ data, selected}) => {
  return (
  <Node isSelected={selected}>
    <Handle type="target" position={Position.Left} id={data.targetHandle} />
    <CustomNode data={data} />
    <Handle type="source" position={Position.Right} id={data.sourceHandle} />
  </Node>
)};

export const ParentNode = ({ data, selected }) => {
  return (
    <Node isSelected={selected}>
      <CustomNode data={data} />
      <Handle type="source" position={Position.Right} id={data.sourceHandle} />
    </Node>
  );};
  
export const CustomNode = ({ data }) => {
  const theme = useTheme()
  const lightTheme = theme.palette.mode === "light"
  const colorPerType =
    data.resourceTrait != 0
      ? IconColors[data.resourceTrait]
      : IconColors[data.resourceType];
  return (
    <NodeWrapper>
      <IconWrapper
        backgroundColor={colorPerType.light}
        borderColor={lightTheme ? colorPerType.light : colorPerType.dark}
      >
        <IconPerType
          resourceTrait={data.resourceTrait}
          color={lightTheme ? colors.white : colorPerType.dark}
          resourceType={data.resourceType}
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
