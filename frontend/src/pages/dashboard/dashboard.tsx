import { Typography, styled } from "@mui/material";
import React from "react";
import { Inventory } from "./inventory/inventory";
import { colors } from "../../style/colors";

export const Wrapper = styled("div")(({ theme }) => ({
  display: "flex",
  height: "100vh",
  marginLeft: "78px",
  flexDirection: "column",
  background: theme.palette.mode === "light" ? colors.white : colors.gray900,
  overflow: "hidden",
}));

const Header = styled("div")(({ theme }) => ({
  display: "flex",
  alignItems: "center",
  padding: "14px 20px 10px 20px",
  borderBottom: `1px solid ${
    theme.palette.mode === "light" ? colors.gray100 : "rgba(255,255,255,0.06)"
  }`,
}));

const Content = styled("div")(() => ({
  display: "flex",
  flex: 1,
  minHeight: 0,
  overflow: "auto",
}));

export const Dashboard = () => {
  return (
    <Wrapper>
      <Header>
        <Typography variant="h6" sx={{ fontSize: 16, fontWeight: 700, m: 0 }}>
          Dashboard
        </Typography>
      </Header>
      <Content>
        <Inventory />
      </Content>
    </Wrapper>
  );
};
