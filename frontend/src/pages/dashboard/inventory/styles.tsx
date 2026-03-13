import { styled } from "@mui/material";

export const IdentityWrapper = styled("div")(({ theme }) => ({
  display: "flex",
  flexWrap: "wrap",
  width: "100%",
  borderRadius: "10px",
  padding: "12px",
  maxWidth: "fit-content",
  height: "fit-content",
}));

export const ResourcesWrapper = styled("div")(() => ({
  display: "flex",
  gap: "12px",
}));

export const LoadingWrapper = styled("div")(() => ({
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
}));
