import { styled } from "@mui/material";
import { colors } from "../../../../style/colors";

export const DataWrapper = styled("div")(() => ({
  display: "flex",
  flexDirection: "column",
  justifyContent: "space-evenly",
  flexWrap: "wrap",
}));

export const Data = styled("div")(() => ({
  display: "flex",
  flexDirection: "column",
  justifyContent: "center",
  textTransform: "uppercase",
  alignItems: "center",
  flexWrap: "wrap",
  fontSize: "14px",
  fontWeight: 600,
  margin: "20px 0",

  "> span": {
    color: colors.orange600,
    fontSize: "30px",
    fontWeight: 600,
  }
}));
