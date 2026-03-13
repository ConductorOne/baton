import { styled } from "@mui/material/styles";
import { IconButton, Paper, Typography } from "@mui/material";
import { Link } from "react-router-dom";
import { colors } from "../../style/colors";

export const ResourceDetailsPanel = styled(Paper)(({ theme }) => ({
  position: "fixed",
  zIndex: 1000,
  padding: "20px",
  maxWidth: "336px",
  width: "336px",
  maxHeight: "70vh",
  overflowY: "auto",
  borderRadius: "12px",
  border: `1px solid ${
    theme.palette.mode === "light" ? colors.gray200 : "rgba(255,255,255,0.08)"
  }`,
  boxShadow:
    theme.palette.mode === "light"
      ? "0px 4px 16px rgba(0,0,0,0.08)"
      : "0px 4px 16px rgba(0,0,0,0.3)",
  backgroundColor:
    theme.palette.mode === "light" ? colors.white : colors.gray800,
}));

export const StyledDiv = styled("div")(() => ({
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  width: "100%",
  marginBottom: "12px",
}));

export const ModalHeader = styled("div")(({ theme }) => ({
  paddingBottom: "12px",
  borderBottom: `1px solid ${
    theme.palette.mode === "light" ? colors.gray200 : "rgba(255,255,255,0.08)"
  }`,
  width: "100%",
  marginBottom: "16px",
}));

export const Details = styled("div")(() => ({
  display: "flex",
  flexDirection: "column",
}));

export const Label = styled(Typography)(({ theme }) => ({
  color: theme.palette.text.secondary,
  fontSize: "12px",
  fontWeight: 500,
  textTransform: "uppercase",
  letterSpacing: "0.3px",
}));

export const Value = styled(Typography)(({ theme }) => ({
  color: theme.palette.text.primary,
  fontSize: "13px",
}));

export const Container = styled("div")(() => ({
  marginBottom: "10px",
}));

export const CloseButton = styled(IconButton)(() => ({
  marginLeft: "5px",
  borderRadius: "6px",
}));

export const StyledLink = styled(Link)(({ theme }) => ({
  textDecoration: "none",
  color: theme.palette.text.primary,
  "&:hover": {
    color: theme.palette.secondary.main,
  },
}));
