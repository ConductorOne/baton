import { Tabs, styled } from "@mui/material";
import { colors } from "../../../style/colors";
import { CardStyleProps } from "./cards/cards";
import { Link } from "react-router-dom";

export type CardSize = "s" | "m" | "l";

export const DefaultWrapper = styled("div", {
  shouldForwardProp: (prop) => prop !== "width" && prop !== "row",
})<{ width?: number; row?: boolean }>(({ theme, width, row }) => ({
  display: "inline-flex",
  flexDirection: row ? "row" : "column",
  borderRadius: "10px",
  background: theme.palette.mode === "light" ? colors.white : colors.gray800,
  border: `1px solid ${
    theme.palette.mode === "light" ? colors.gray200 : "rgba(255,255,255,0.08)"
  }`,
  padding: "12px",
  width: "100%",
  maxWidth: width ? `${width}px` : "auto",
  height: "max-content",
  flexWrap: "wrap",
}));

export const DefaultContainer = styled("div")(({ theme }) => ({
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
  flexDirection: "column",
  borderRadius: "8px",
  border:
    theme.palette.mode === "light"
      ? `1px solid ${colors.gray100}`
      : "1px solid rgba(255,255,255,0.06)",
}));

export const DataWrapper = styled(Link, {
  shouldForwardProp: (prop) =>
    prop !== "isColumn" &&
    prop !== "size" &&
    prop !== "withoutBackground" &&
    prop !== "noBorder" &&
    prop !== "fullWidth" &&
    prop !== "withMargin" &&
    prop !== "topRadius",
})<CardStyleProps>(
  ({
    theme,
    isColumn,
    topRadius,
    noBorder,
    withoutBackground,
    size,
    fullWidth,
    withMargin,
  }) => ({
    display: "flex",
    textDecoration: "none",
    justifyContent: isColumn ? "center" : "space-evenly",
    alignItems: "center",
    flexDirection: isColumn ? "column" : "row",
    borderRadius: topRadius ? "8px 8px 0 0" : "8px",
    border: noBorder
      ? "none"
      : theme.palette.mode === "light"
      ? `1px solid ${colors.gray100}`
      : "1px solid rgba(255,255,255,0.06)",
    background:
      theme.palette.mode === "light"
        ? withoutBackground
          ? colors.white
          : colors.gray50
        : withoutBackground
        ? colors.gray800
        : colors.gray900,
    padding: "16px",
    minWidth: size === "l" ? "auto" : "180px",
    width: fullWidth ? "100%" : "auto",
    margin: withMargin ? "0 6px" : 0,

    "&:first-of-type": {
      margin: "0",
      marginRight: withMargin ? "6px" : 0,
    },
    "&:last-of-type": {
      margin: "0",
      marginLeft: withMargin ? "6px" : 0,
    },
  })
);

export const LayoutWrapper = styled("div", {
  shouldForwardProp: (prop) => prop !== "isColumn",
})<{ isColumn?: boolean }>(({ isColumn }) => ({
  display: "flex",
  flexDirection: isColumn ? "column" : "row",
  flexWrap: "wrap",
  padding: "16px",
  width: "100%",
}));

export const StyledTabs = styled(Tabs)(() => ({
  ".MuiTabs-indicator": {
    display: "none",
  },
}));
