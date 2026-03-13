import MuiDrawer from "@mui/material/Drawer";
import { styled } from "@mui/material/styles";
import { IconButton, ListItemButton, List, Typography } from "@mui/material";
import { colors } from "../../../style/colors";

export const ExplorerLayout = styled("div", {
  shouldForwardProp: (prop) => prop !== "sidebarOpen",
})<{ sidebarOpen?: boolean }>(({ sidebarOpen }) => {
  const navWidth = 78;
  const sidebarWidth = sidebarOpen ? 270 : 0;
  const totalOffset = navWidth + sidebarWidth;
  return {
    display: "flex",
    flexDirection: "column",
    width: `calc(100vw - ${totalOffset}px)`,
    height: "100vh",
    marginLeft: `${totalOffset}px`,
    transition: "margin-left 0.2s ease, width 0.2s ease",
  };
});

export const TreeWrapper = styled("div")(() => ({
  flex: 1,
  minHeight: 0,
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
  padding: "16px",
}));

export const ResourcesListWrapper = styled(List)(() => ({
  width: "100%",
  marginTop: "16px",
}));

export const Sidebar = styled(MuiDrawer)(({ theme }) => ({
  "& .MuiDrawer-paper": {
    maxWidth: "270px",
    width: "100%",
    display: "flex",
    backgroundColor:
      theme.palette.mode === "light" ? colors.gray900 : colors.gray950,
    color: colors.gray50,
    boxShadow: "2px 0 8px rgba(0,0,0,0.08)",
    marginLeft: "78px",
    padding: "16px",
  },
}));

export const ResourceLabel = styled(ListItemButton)(({ theme }) => ({
  borderRadius: "6px",
  color: colors.gray100,
  padding: "6px 12px",
  marginBottom: "1px",
  p: {
    fontSize: "13px",
  },
  "&:hover": {
    backgroundColor: "rgba(255,255,255,0.06)",
  },
  "&.Mui-selected": {
    backgroundColor: "rgba(155,237,117,0.12)",
    "> p": {
      fontWeight: 600,
      color: colors.batonGreen300,
    },
    "&:hover": {
      backgroundColor: "rgba(155,237,117,0.16)",
    },
  },
}));

export const SidebarHeader = styled("div")(() => ({
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  padding: "0 0 12px 12px",
}));

export const StyledButton = styled(IconButton)(() => ({
  borderRadius: "6px",
  border: "1px solid rgba(255,255,255,0.12)",
  color: colors.gray300,
  "&:hover": {
    backgroundColor: "rgba(255,255,255,0.06)",
  },
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
  ({ backgroundColor, borderColor }) => ({
    backgroundColor: backgroundColor,
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
  padding: "0 16px",
  color: colors.gray400,
}));

export const EntitlementNumberLabel = styled("span")(({ theme }) => ({
  backgroundColor: theme.palette.secondary.main,
  marginRight: "5px",
  color: colors.white,
  borderRadius: "1000px",
  padding: "1px 5px",
  fontSize: "8px",
}));

export const SelectedEntitlementWrapper = styled("div")(() => ({
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
}));

export const Node = styled("div", {
  shouldForwardProp: (prop) => prop !== "isSelected",
})<{ isSelected: boolean }>(({ theme, isSelected }) => ({
  backgroundColor: isSelected
    ? theme.palette.mode === "light"
      ? colors.batonGreen100
      : "rgba(155,237,117,0.08)"
    : theme.palette.mode === "light"
    ? colors.white
    : colors.gray800,
  border: isSelected
    ? `1.5px solid ${
        theme.palette.mode === "light"
          ? colors.batonGreen600
          : colors.batonGreen500
      }`
    : `1px solid ${
        theme.palette.mode === "light" ? colors.gray200 : "rgba(255,255,255,0.08)"
      }`,
  display: "flex",
  padding: "12px 14px",
  alignItems: "center",
  borderRadius: "10px",
  maxWidth: "300px",
  minWidth: "200px",
  boxShadow: isSelected
    ? "none"
    : theme.palette.mode === "light"
    ? "0px 1px 3px rgba(0,0,0,0.06)"
    : "none",

  color:
    theme.palette.mode === "light" ? colors.gray900 : colors.gray50,
  span: {
    color:
      theme.palette.mode === "light" ? colors.gray600 : colors.gray300,
  },

  ".react-flow__handle": {
    background: `${colors.batonGreen600} !important`,
  },
}));
