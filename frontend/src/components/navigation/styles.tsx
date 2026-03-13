import MuiDrawer from "@mui/material/Drawer";
import { styled } from "@mui/material/styles";
import { ListItemButton } from "@mui/material";
import { Link } from "react-router-dom";
import { colors } from "../../style/colors";

export const StyledDrawer = styled(MuiDrawer)(({ theme }) => ({
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  "& .MuiDrawer-paper": {
    backgroundColor:
      theme.palette.mode === "light" ? colors.gray900 : colors.gray950,
    boxShadow: "1px 0 4px rgba(0,0,0,0.08)",
    boxSizing: "border-box",
    alignItems: "center",
    padding: "0 20px 20px 20px",
    maxWidth: "78px",
    width: "100%",
    color: colors.gray50,
    zIndex: 99999,
    borderRight: "none",
  },
}));

export const CloseButton = styled(ListItemButton)(() => ({
  margin: "8px 0",
}));

export const IconWrapper = styled("div", {
  shouldForwardProp: (prop) => prop !== "isSelected",
})<{ isSelected?: boolean }>(({ theme, isSelected }) => ({
  display: "flex",
  padding: "8px",
  alignItems: "center",
  justifyContent: "center",
  borderRadius: "6px",
  border: isSelected
    ? `1px solid ${colors.batonGreen500}`
    : "1px solid transparent",
  background: isSelected
    ? "rgba(155,237,117,0.12)"
    : "transparent",
  transition: "all 0.15s ease",
  "&:hover": {
    background: isSelected
      ? "rgba(155,237,117,0.16)"
      : "rgba(255,255,255,0.06)",
  },
}));

export const NavWrapper = styled("div")(() => ({
  display: "flex",
  alignItems: "center",
  justifyContent: "start",
  flexDirection: "column",
  height: "100%",
}));

export const StyledLink = styled(Link)(() => ({
  textDecoration: "none",
  color: colors.gray200,
}));
