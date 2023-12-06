import React from "react";

import { useTheme } from "@mui/material";
import DashboardOutlinedIcon from "@mui/icons-material/DashboardOutlined";
import { useLocation } from "react-router-dom";
import { IconWrapper, StyledLink } from "../styles";

export const DashboardButton = () => {
  const location = useLocation();
  const theme = useTheme();
  const isSelected = location.pathname === "/dashboard";
  const iconColor = isSelected
    ? theme.palette.secondary.main
    : theme.palette.mode === "light"
    ? theme.palette.primary.dark
    : theme.palette.primary.contrastText;

  return (
    <StyledLink sx={{ margin: "21px 0" }} to="/dashboard">
      <IconWrapper isSelected={isSelected}>
        <DashboardOutlinedIcon htmlColor={iconColor} />
      </IconWrapper>
    </StyledLink>
  );
};
