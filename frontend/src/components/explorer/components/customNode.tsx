import React from "react";
import { Handle, Position } from "reactflow";
import { Typography, useTheme } from "@mui/material";
import ChevronRightIcon from "@mui/icons-material/ChevronRight";
import VpnKeyOutlinedIcon from "@mui/icons-material/VpnKeyOutlined";
import { normalizeString } from "../../../common/helpers";
import {
  IconWrapper,
  Node,
  NodeInfoWrapper,
  NodeWrapper,
} from "../styles/styles";
import { colors } from "../../../style/colors";
import { IconPerType, IconColors } from "../../icons/resourceTypeIcon";

export const ChildNode = ({ data, selected }) => {
  return (
    <Node isSelected={selected}>
      <Handle type="target" position={Position.Left} id={data.targetHandle} />
      <CustomNode data={data} />
    </Node>
  );
};

export const ExpandableGrantNode = ({ data, selected }) => {
  return (
    <Node isSelected={selected}>
      <Handle type="target" position={Position.Left} id={data.targetHandle} />
      <CustomNode data={data} />
      <Handle type="source" position={Position.Right} id={data.sourceHandle} />
    </Node>
  );
};

export const ParentNode = ({ data, selected }) => {
  return (
    <Node isSelected={selected}>
      <CustomNode data={data} />
      <Handle type="source" position={Position.Right} id={data.sourceHandle} />
    </Node>
  );
};

export const EntitlementNode = ({ data, selected }) => {
  const theme = useTheme();
  const lightTheme = theme.palette.mode === "light";
  return (
    <Node isSelected={selected}>
      <Handle type="target" position={Position.Left} id={data.targetHandle} />
      <NodeWrapper>
        <IconWrapper
          backgroundColor={lightTheme ? colors.indigo700 : colors.black}
          borderColor={lightTheme ? colors.indigo700 : colors.indigo400}
        >
          <VpnKeyOutlinedIcon sx={{ color: lightTheme ? colors.white : colors.indigo400, fontSize: 16 }} />
        </IconWrapper>
        <NodeInfoWrapper>
          <Typography>{data.label}</Typography>
          <Typography variant="caption">entitlement</Typography>
        </NodeInfoWrapper>
      </NodeWrapper>
      <Handle type="source" position={Position.Right} id={data.sourceHandle} />
    </Node>
  );
};

export const AggregateNode = ({ data, selected }) => {
  return (
    <Node isSelected={selected}>
      <Handle type="target" position={Position.Left} id={data.targetHandle} />
      <CustomNode data={data} />
      <ChevronRightIcon sx={{ ml: "auto", opacity: 0.6, flexShrink: 0 }} />
    </Node>
  );
};

export const CustomNode = ({ data }) => {
  const theme = useTheme();
  const lightTheme = theme.palette.mode === "light";
  const defaultColor = { light: colors.purple800, dark: colors.purple400 };
  const colorPerType =
    data.resourceTrait !== 0
      ? IconColors[data.resourceTrait] || defaultColor
      : IconColors[data.resourceType] || defaultColor;
  return (
    <NodeWrapper>
      <IconWrapper
        backgroundColor={lightTheme ? colorPerType.light : colors.black}
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
