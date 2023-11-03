import React from "react";

import { styled } from "@mui/material";
import { BatonLogo } from "../../icons/icons";

const StyledLogo = styled('div')(({ theme }) => ({
  display: "flex",
  width: "48px",
  height: "48px",
  padding: "8px 12px 8px 13px",
  justifyContent: "center",
  alignItems: "center",
  borderRadius: "0px 0px 7px 7px",
  background: theme.palette.mode === "light" ? theme.palette.secondary.main : theme.palette.primary.dark,    
}));

export const Logo = () => {
  return (
    <StyledLogo>
      <BatonLogo />
    </StyledLogo>
  );
};
