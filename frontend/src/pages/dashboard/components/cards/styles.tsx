import { styled } from "@mui/material";
import { colors } from "../../../../style/colors";

export const SizeMap = {
  s: {
    count: {
      fontSize: "30px",
      fontWeight: 600,
    },
    label: {
      fontSize: "16px",
    },
  },
  m: {
    count: {
      fontSize: "60px",
      fontWeight: 600,
    },
    label: {
      fontSize: "18px",
    },
  },
  l: {
    count: {
      fontSize: "72px",
      fontWeight: 700,
    },
    label: {
      fontSize: "20px",
    },
  },
};

export const Label = styled("span", {
  shouldForwardProp: (prop) => prop !== "size",
})<{ size?: any }>(({ size, theme }) => ({
  textTransform: "uppercase",
  fontSize: size.fontSize,
  color: theme.palette.mode === "light" ? colors.batonGreen1000 : colors.gray200,
}));

export const Count = styled("span", {
  shouldForwardProp: (prop) => prop !== "size",
})<{ size?: any }>(({ size, theme }) => ({
  justifyContent: "center",
  fontSize: size.fontSize,
  fontWeight: size.fontWeight,
  color: theme.palette.mode === "light" ? colors.batonGreen800 : colors.batonGreen500,
}));


export const Score = styled("div")(({ theme }) => ({
  justifyContent: "center",
  fontSize: "24px",
  fontWeight: 500,
  color: colors.batonGreen600,

  "> span": {
    color: theme.palette.mode === "light" ? colors.orange600 : colors.orange500,
    fontSize: "72px",
    fontWeight: 700,
  }
}));
