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
  borderRadius: "16px",
  background: theme.palette.mode === "light" ? colors.white : colors.gray700,
  padding: "16px",
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
  borderRadius: "16px",
  border:
    theme.palette.mode === "light" ? `1px solid ${colors.gray200}` : "none",
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
    justifyContent: isColumn ? "center" : "space-between",
    alignItems: "center",
    flexDirection: isColumn ? "column" : "row",
    borderRadius: topRadius ? "16px 16px 0 0" : "16px",
    border: noBorder
      ? "none"
      : theme.palette.mode === "light"
      ? `1px solid ${colors.gray200}`
      : "none",
    background:
      theme.palette.mode === "light"
        ? withoutBackground
          ? colors.white
          : colors.gray50
        : withoutBackground
        ? colors.gray900
        : colors.black,
    padding: "25px",
    minWidth: size === "l" ? "350px" : "210px",
    width: fullWidth ? "100%" : "auto",
    margin: withMargin ? "0 8px" : 0,

    "&:first-of-type": {
      margin: "0",
      marginRight: withMargin ? "8px" : 0,
    },
    "&:last-of-type": {
      margin: "0",
      marginLeft: withMargin ? "8px" : 0,
    },
  })
);

export const LayoutWrapper = styled("div", {
  shouldForwardProp: (prop) => prop !== "isColumn",
})<{ isColumn?: boolean }>(({ isColumn }) => ({
  display: "flex",
  flexDirection: isColumn ? "column" : "row",
  flexWrap: "wrap",
  padding: "25px",
  width: "100%",
}));

export const StyledTabs = styled(Tabs)(() => ({
  ".MuiTabs-indicator": {
    display: "none",
  },
}));
