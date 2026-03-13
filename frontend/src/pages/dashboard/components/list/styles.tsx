import { List, styled } from "@mui/material";
import { Link } from "react-router-dom";
import { colors } from "../../../../style/colors";

export const StyledList = styled(List)(({ theme }) => ({
  background: theme.palette.mode === "light" ? colors.white : colors.gray800,
  width: "100%",
}));

export const ListItem = styled(Link)(({ theme }) => ({
  display: "flex",
  justifyContent: "space-between",
  textDecoration: "none",
  fontSize: "13px",
  padding: "6px 12px",
  color: theme.palette.text.primary,
  alignItems: "center",
  borderRadius: "4px",
  "&:hover": {
    backgroundColor:
      theme.palette.mode === "light" ? colors.gray50 : "rgba(255,255,255,0.04)",
  },
}));

export const Count = styled("div")(() => ({
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  fontWeight: 600,
  fontSize: "13px",

  span: {
    marginRight: "5px",
  },
}));

export const ButtonWrapper = styled("div")(({ theme }) => ({
  display: "flex",
  padding: "6px 10px",
  justifyContent: "center",
  alignItems: "center",
  gap: "6px",
  borderRadius: "6px",
  border: `1px solid ${
    theme.palette.mode === "light" ? colors.gray200 : "rgba(255,255,255,0.12)"
  }`,
}));

export const StyledListItem = styled(Link)(({ theme }) => ({
  fontSize: "13px",
  color: theme.palette.text.secondary,
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
  padding: "6px",
  textDecoration: "none",
  "&:hover": {
    color: theme.palette.secondary.main,
  },
}));

export const Title = styled("p")(({ theme }) => ({
  padding: "0 12px",
  color: theme.palette.text.secondary,
  fontSize: "11px",
  fontWeight: 600,
  textTransform: "uppercase",
  letterSpacing: "0.5px",
}));
