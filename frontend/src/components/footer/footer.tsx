import React from "react"
import { styled } from "@mui/material/styles";
import { ExplorerLogo } from "../icons/icons";

const FooterWrapper = styled("div")(() => ({
  position: "fixed",
  bottom: "0",
  left: "72px",
}));

const Footer = () => (
  <FooterWrapper>
    <ExplorerLogo color="secondary" />
  </FooterWrapper>
);

export default Footer