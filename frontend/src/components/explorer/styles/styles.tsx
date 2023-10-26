import MuiDrawer from "@mui/material/Drawer";
import { styled } from "@mui/material/styles";
import { IconButton, ListItemButton, List, Typography } from "@mui/material";

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
    backgroundColor: theme.palette.primary.dark,
    marginLeft: "72px",
    padding: "20px",
  },
}));

export const ResourceLabel = styled(ListItemButton)(() => ({
  borderRadius: "8px",
  color: "white",
  fontSize: "14px",
  padding: "8px 16px",
  marginBottom: "2px",
  marginRight: "0px",
  "&:hover": {
    backgroundColor: "rgba(255, 255, 255, 0.10)",
  },
  "&.Mui-selected": {
    backgroundColor: "rgba(255, 255, 255, 0.30)",
  },
}));

export const SidebarHeader = styled("div")(() => ({
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  borderBottom: "1px solid rgba(255, 255, 255, 0.10)",
  padding: "0 0 20px 16px",
}));

export const StyledButton = styled(IconButton)(() => ({
  border: "1px solid rgba(255, 255, 255, 0.30)",
  borderRadius: "8px",
  boxShadow: "0px 1px 2px 0px rgba(16, 24, 40, 0.05)",
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

export const IconWrapper = styled("div")(({ theme }) => ({
  backgroundColor: theme.palette.primary.contrastText,
  borderRadius: "1000px",
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
  marginRight: "10px",
  padding: "6px",
  border: "1px solid #E0E0E1",
}));

export const EmptyResourceLabel = styled(Typography)(() => ({
  padding: "0 20px",
}));

export const EntitlementNumberLabel = styled("span")(({ theme }) => ({
  backgroundColor: theme.palette.primary.main,
  marginRight: "5px",
  color: theme.palette.primary.contrastText,
  borderRadius: "1000px",
  padding: "1px 4px",
  fontSize: "8px",
}));

export const SelectedEntitlementWrapper = styled("div")(() => ({
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
}));
