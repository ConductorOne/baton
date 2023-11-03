import MuiDrawer from "@mui/material/Drawer";
import { styled } from "@mui/material/styles";
import { IconButton, ListItemButton, List, Typography } from "@mui/material";
import { colors } from "../../../style/colors"

export const TreeWrapper = styled("div")(() => ({
  width: "100vw",
  height: "100vh",
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
  padding: "20px"
}));

export const ResourcesListWrapper = styled(List)(() => ({
  width: "100%",
  marginTop: "20px",
}));

export const Sidebar = styled(MuiDrawer)(({ theme }) => ({
  "& .MuiDrawer-paper": {
    maxWidth: "270px",
    width: "100%",
    display: "flex",
    backgroundColor:
      theme.palette.mode === "light"
        ? theme.palette.primary.main
        : colors.gray900,
    color: theme.palette.primary.contrastText,
    boxShadow:
      "2px 0px 16px 0px rgba(0, 0, 0, 0.02), 3px 0px 8px 0px rgba(0, 0, 0, 0.03)",
    marginLeft: "78px",
    padding: "20px",
  },
}));

export const ResourceLabel = styled(ListItemButton)(({ theme }) => ({
  borderRadius: "8px",
  color: "white",
  padding: "8px 16px",
  marginBottom: "2px",
  marginRight: "0px",
  p: {
    fontSize: "14px",
  },
  "&:hover": {
    backgroundColor:
      theme.palette.mode === "light"
        ? theme.palette.secondary.light
        : theme.palette.primary.dark,
  },
  "&.Mui-selected": {
    backgroundColor:
      theme.palette.mode === "light"
        ? theme.palette.secondary.light
        : theme.palette.primary.dark,
    "> p": {
      fontWeight: "bolder",
    },
    "&:hover": {
      backgroundColor:
        theme.palette.mode === "light"
          ? theme.palette.secondary.light
          : theme.palette.primary.dark,
    },
  },
}));

export const SidebarHeader = styled("div")(() => ({
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  padding: "0 0 20px 16px",
}));

export const StyledButton = styled(IconButton)(({ theme }) => ({
  borderRadius: "8px",
  boxShadow: "0px 1px 2px 0px rgba(16, 24, 40, 0.05)",
  marginLeft: "5px",
}));

export const NodeInfoWrapper = styled("div")(() => ({
  display: "flex",
  flexDirection: "column",
}));

export const NodeWrapper = styled("div")(() => ({
  display: "flex",
  alignItems: "center",
  width: "100%",
}));

export const IconWrapper = styled("div", {
  shouldForwardProp: (prop) =>
    prop !== "backgroundColor" && prop !== "borderColor",
})<{ backgroundColor?: string; borderColor?: string }>(
  ({ backgroundColor, borderColor, theme }) => ({
    backgroundColor: theme.palette.mode === "light" ? backgroundColor : colors.black,
    borderRadius: "1000px",
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    marginRight: "10px",
    padding: "6px",
    border: `1px solid ${borderColor}`,
  })
);

export const EmptyResourceLabel = styled(Typography)(() => ({
  padding: "0 20px",
}));

export const EntitlementNumberLabel = styled("span")(({ theme }) => ({
  backgroundColor: theme.palette.secondary.main,
  marginRight: "5px",
  color: colors.white,
  borderRadius: "1000px",
  padding: "1px 5px",
  fontSize: "8px",
}));

export const SelectedEntitlementWrapper = styled("div")(({ theme }) => ({
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
}));

export const Node = styled("div", {
  shouldForwardProp: (prop) => prop !== "isSelected",
})<{ isSelected: boolean }>(({ theme, isSelected }) => ({
  backgroundColor: isSelected
    ? theme.palette.mode === "light"
      ? colors.batonGreen200
      : colors.batonGreen900
    : theme.palette.mode === "light"
    ? colors.white
    : colors.gray700,
  border: isSelected
    ? `1.2px solid ${
        theme.palette.mode === "light"
          ? colors.batonGreen700
          : colors.batonGreen500
      }`
    : `1.2px solid ${
        theme.palette.mode === "light"
          ? colors.white
          : colors.gray700
      }`,
  display: "flex",
  padding: " 16px 16px 16px 12px",
  alignItems: "center",
  borderRadius: "12px",
  maxWidth: "300px",
  minWidth: "200px",
  boxShadow: isSelected
    ? "none"
    : "0px 2px 4px -2px rgba(16, 24, 40, 0.06), 0px 4px 8px -2px rgba(16, 24, 40, 0.10)",

  color:
    theme.palette.mode === "light"
      ? colors.batonGreen1000
      : colors.batonGreen100,
  span: {
    color:
      theme.palette.mode === "light"
        ? colors.batonGreen1000
        : colors.batonGreen200
  },

  ".react-flow__handle": {
    background: `${colors.batonGreen600} !important`,
  },
}));
