import React from "react";
import PersonOutlinedIcon from "@mui/icons-material/PersonOutlined";
import GroupsOutlinedIcon from "@mui/icons-material/GroupsOutlined";
import LibraryBooksOutlinedIcon from "@mui/icons-material/LibraryBooksOutlined";
import WidgetsOutlinedIcon from "@mui/icons-material/WidgetsOutlined";
import BadgeOutlinedIcon from "@mui/icons-material/BadgeOutlined";
import LockOutlinedIcon from "@mui/icons-material/LockOutlined";
import IntegrationInstructionsOutlinedIcon from "@mui/icons-material/IntegrationInstructionsOutlined";
import MapsHomeWorkOutlinedIcon from "@mui/icons-material/MapsHomeWorkOutlined";
import AccountTreeOutlinedIcon from "@mui/icons-material/AccountTreeOutlined";
import PasswordOutlinedIcon from "@mui/icons-material/PasswordOutlined";
import { colors } from "../../style/colors";

export const IconColors = {};

export const IconPerType = ({
  resourceTrait = 0,
  color = colors.white,
  resourceType,
}) => {
  switch (resourceTrait) {
    case 1:
      IconColors[1] = {
        dark: colors.batonGreen300,
        light: colors.batonGreen600,
      };
      return <PersonOutlinedIcon htmlColor={color} />;
    case 2:
      IconColors[2] = {
        dark: colors.batonGreen400,
        light: colors.batonGreen800,
      };
      return <GroupsOutlinedIcon htmlColor={color} />;
    case 3:
      IconColors[3] = {
        dark: colors.batonGreen600,
        light: colors.batonGreen1000,
      };
      return <BadgeOutlinedIcon htmlColor={color} />;
    case 4:
      IconColors[4] = {
        dark: colors.teal400,
        light: colors.teal700,
      };
      return <WidgetsOutlinedIcon htmlColor={color} />;
    case 0:
      switch (true) {
        case resourceType.includes("vault"):
          IconColors[resourceType] = {
            dark: colors.yellow500,
            light: colors.yellow800,
          };
          return <LockOutlinedIcon htmlColor={color} />;

        case resourceType.includes("repository"):
        case resourceType.includes("repo"):
          IconColors[resourceType] = {
            dark: colors.white,
            light: colors.black,
          };
          return <IntegrationInstructionsOutlinedIcon htmlColor={color} />;

        case resourceType.includes("org"):
        case resourceType.includes("organisation"):
        case resourceType.includes("organization"):
          IconColors[resourceType] = {
            dark: colors.indigo400,
            light: colors.indigo700,
          };
          return <MapsHomeWorkOutlinedIcon htmlColor={color} />;

        case resourceType.includes("integration"):
          IconColors[resourceType] = {
            dark: colors.purple300,
            light: colors.purple700,
          };
          return <AccountTreeOutlinedIcon htmlColor={color} />;

        case resourceType.includes("credential"):
          IconColors[resourceType] = {
            dark: colors.orange400,
            light: colors.orange700,
          };
          return <PasswordOutlinedIcon htmlColor={color} />;
        default:
          IconColors[resourceType] = {
            dark: colors.purple400,
            light: colors.purple800,
          };
          return <LibraryBooksOutlinedIcon htmlColor={color} />;
      }
    default:
      IconColors[resourceType] = {
        dark: colors.purple400,
        light: colors.purple800,
      };
      return <LibraryBooksOutlinedIcon htmlColor={color} />;
  }
};
