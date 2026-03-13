import React from "react";

import DashboardOutlinedIcon from "@mui/icons-material/DashboardOutlined";
import { useLocation } from "react-router-dom";
import { IconWrapper, StyledLink } from "../styles";
import { colors } from "../../../style/colors";

export const DashboardButton = () => {
  const location = useLocation();
  const isSelected = location.pathname === "/dashboard";
  const iconColor = isSelected ? colors.batonGreen400 : colors.gray400;

  return (
    <StyledLink sx={{ margin: "21px 0" }} to="/dashboard">
      <IconWrapper isSelected={isSelected}>
        <DashboardOutlinedIcon htmlColor={iconColor} />
      </IconWrapper>
    </StyledLink>
  );
};
