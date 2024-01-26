import React from "react"
import { styled } from "@mui/material/styles";
import { ExplorerLogo, C1Logo } from "../icons/icons";
import { Divider } from "@mui/material";
import { colors } from "../../style/colors";

const FooterWrapper = styled("div")(() => ({
  position: "absolute",
  top: "0",
  right: "0",
  display: "flex",
  alignItems: "center",
  padding: "20px",

  hr: {
    margin: "0 8px",
    backgroundColor: colors.gray200,
  },
}));

const StyledSpan = styled("span")(({ theme }) => ({
  fontSize: "6px",
  marginRight: "5px",
  color: colors.gray400
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