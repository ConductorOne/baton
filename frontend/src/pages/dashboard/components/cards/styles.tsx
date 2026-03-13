import { styled } from "@mui/material";
import { colors } from "../../../../style/colors";

export const SizeMap = {
  s: {
    count: {
      fontSize: "28px",
      fontWeight: 600,
    },
    label: {
      fontSize: "13px",
    },
  },
  m: {
    count: {
      fontSize: "40px",
      fontWeight: 600,
    },
    label: {
      fontSize: "14px",
    },
  },
  l: {
    count: {
      fontSize: "48px",
      fontWeight: 700,
    },
    label: {
      fontSize: "16px",
    },
  },
};

export const Label = styled("span", {
  shouldForwardProp: (prop) => prop !== "size" && prop !== "marginRight",
})<{ size?: any; marginRight?: boolean }>(({ size, theme, marginRight }) => ({
  textTransform: "uppercase",
  fontSize: size.fontSize,
  letterSpacing: "0.3px",
  color: theme.palette.text.secondary,
  marginBottom: marginRight ? "0" : "6px",
  marginRight: marginRight ? "16px" : "0",
}));

export const Count = styled("span", {
  shouldForwardProp: (prop) => prop !== "size",
})<{ size?: any }>(({ size, theme }) => ({
  justifyContent: "center",
  fontSize: size.fontSize,
  fontWeight: size.fontWeight,
  color:
    theme.palette.mode === "light"
      ? colors.batonGreen700
      : colors.batonGreen500,
}));

export const Score = styled("div")(({ theme }) => ({
  justifyContent: "center",
  fontSize: "24px",
  fontWeight: 500,
  color: colors.batonGreen600,

  "> span": {
    color:
      theme.palette.mode === "light" ? colors.orange600 : colors.orange500,
    fontSize: "48px",
    fontWeight: 700,
  },
}));
