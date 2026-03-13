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
  background: "rgba(255,255,255,0.08)",
}));

export const Logo = () => {
  return (
    <StyledLogo>
      <BatonLogo />
    </StyledLogo>
  );
};
