import { styled } from "@mui/material/styles";
import MuiDrawer from "@mui/material/Drawer";
import { IconButton, Typography } from "@mui/material";

export const ResourceDetailsDrawer = styled(MuiDrawer)(({ theme }) => ({
  "& .MuiDrawer-paper": {
    padding: "20px",
    margin: "20px",
    display: "flex",
    maxWidth: "336px",
    width: "100%",
    height: "auto",
    flexDirection: "column",
    alignItems: "flex-start",
    borderRadius: "16px",
    border: "1px solid #EAECF0",
    boxShadow:
      "2px 0px 16px 0px rgba(0, 0, 0, 0.02), 3px 0px 8px 0px rgba(0, 0, 0, 0.03)",
  },
}));

export const StyledDiv = styled("div")(() => ({
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  width: "100%",
  marginBottom: "15px",
}));

export const ModalHeader = styled("div")(() => ({
  paddingBottom: "15px",
  borderBottom: "1px solid #EAECF0",
  width: "100%",
  marginBottom: "20px"
}))

export const Details = styled("div")(() => ({
  display: "flex",
  flexDirection: "column",
}));

export const Label = styled(Typography)(({ theme }) => ({
  color: theme.palette.secondary.main,
}));

export const Value = styled(Typography)(({ theme }) => ({
  color: theme.palette.secondary.dark,
}));

export const Container = styled("div")(() => ({
  marginBottom: "10px",
}));

export const CloseButton = styled(IconButton)(() => ({
   border: "1px solid #D0D5DD", 
   borderRadius: "8px"
}));
