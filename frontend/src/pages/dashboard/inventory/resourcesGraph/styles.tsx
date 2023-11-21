import { styled } from "@mui/material/styles";
import { colors } from "../../../../style/colors";

export const ChartWrapper = styled("div")(({ theme }) => ({
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  height: "100%",
  width: "100%",
  background: theme.palette.mode === "light" ? colors.gray50 : colors.black,
  borderRadius:" 16px 16px 0px 0px",
}));

export const DataWrapper = styled("div")(() => ({
  display: "flex",
  justifyContent: "space-evenly",
  flexWrap: "wrap",
}));