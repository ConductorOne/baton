import MuiDrawer from "@mui/material/Drawer";
import { styled } from "@mui/material/styles";
import { ListItemButton } from "@mui/material";
import { Link } from "react-router-dom";

export const StyledDrawer = styled(MuiDrawer)(({ theme }) => ({
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  "& .MuiDrawer-paper": {
    backgroundColor: theme.palette.primary.main,
    boxShadow:
      "2px 0px 16px 0px rgba(0, 0, 0, 0.02), 3px 0px 8px 0px rgba(0, 0, 0, 0.03)",
    boxSizing: "border-box",
    alignItems: "center",
    padding: "0 20px 20px 20px",
    maxWidth: "78px",
    width: "100%",
    color: theme.palette.primary.contrastText,
    zIndex: 99999,
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
    ? `1px solid ${theme.palette.secondary.main}`
    : theme.palette.primary.main,
  background: isSelected
    ? theme.palette.mode === "light"
      ? theme.palette.secondary.light
      : theme.palette.primary.dark
    : theme.palette.primary.main,
}));

export const NavWrapper = styled("div")(() => ({
    display: "flex",
    alignItems: "center",
    justifyContent: "start",
    flexDirection: "column",
    height: "100%",
}))

export const StyledLink = styled(Link)(({ theme }) => ({
  textDecoration: "none",
  color:
    theme.palette.mode === "light"
      ? theme.palette.primary.dark
      : theme.palette.primary.contrastText,
}));
