import { styled } from "@mui/material/styles";

export const ChartWrapper = styled("div")(({ theme }) => ({
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  height: "100%",
  width: "100%",
  flexDirection: "column",
  padding: "12px 0",
}));

export const DataWrapper = styled("div")(() => ({
  display: "flex",
  justifyContent: "space-evenly",
  flexWrap: "wrap",
}));
