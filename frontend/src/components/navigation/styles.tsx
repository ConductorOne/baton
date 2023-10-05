import MuiDrawer from "@mui/material/Drawer";
import { styled } from "@mui/material/styles";
import { ListItemButton } from "@mui/material";

export const Logo = styled("div")(() => ({
  width: "48px",
  height: "48px",
}));

export const StyledDrawer = styled(MuiDrawer)(({ theme }) => ({
  display: "flex",
  alignItems: "center",
  "& .MuiDrawer-paper": {
    backgroundColor: theme.palette.primary.main,
    boxSizing: "border-box",
    alignItems: "center",
    padding: "20px",
    maxWidth: "72px",
    width: "100%",
  },
}));

export const CloseButton = styled(ListItemButton)(() => ({
  margin: "8px 0",
}));
