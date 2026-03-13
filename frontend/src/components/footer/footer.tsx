import React from "react"
import { styled } from "@mui/material/styles";
import { ExplorerLogo, C1Logo } from "../icons/icons";
import { Divider } from "@mui/material";
import { colors } from "../../style/colors";

const FooterWrapper = styled("div")(() => ({
  position: "fixed",
  top: "12px",
  right: "12px",
  display: "flex",
  alignItems: "center",
  padding: "6px 10px",
  borderRadius: "6px",
  backgroundColor: "rgba(255,255,255,0.8)",
  backdropFilter: "blur(4px)",
  zIndex: 100,
  pointerEvents: "none",

  hr: {
    margin: "0 6px",
    backgroundColor: colors.gray200,
  },
}));

const StyledSpan = styled("span")(() => ({
  fontSize: "6px",
  marginRight: "5px",
  color: colors.gray400,
}));

const Footer = () => (
  <FooterWrapper>
    <ExplorerLogo />
    <Divider orientation="vertical" flexItem />
    <StyledSpan>by</StyledSpan>
    <C1Logo />
  </FooterWrapper>
);

export default Footer
