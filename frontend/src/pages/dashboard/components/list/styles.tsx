import { List, styled } from "@mui/material";
import { Link } from "react-router-dom";
import { colors } from "../../../../style/colors";

export const StyledList = styled(List)(({ theme }) => ({
  width: "100%",
}));

export const ListItem = styled(Link)(({ theme }) => ({
  display: "flex",
  justifyContent: "space-between",
  textDecoration: "none",
  fontSize: "18px",
  padding: "8px 16px",
  color: theme.palette.mode === "light" ? colors.batonGreen1000 : colors.white,
  alignItems: "center",
}));

export const Count = styled("div")(() => ({
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  fontWeight: 600,
  fontSize: "18px",
}));

export const ButtonWrapper = styled('div')(({}) => ({
  display: "flex",
  padding:"8px 14px",
  justifyContent: "center",
  alignItems: "center",
  gap: "8px",
  borderRadius: "8px",
  border:`1px solid ${colors.gray300}`,
  boxShadow: "0px 1px 2px 0px rgba(16, 24, 40, 0.05)",
}))