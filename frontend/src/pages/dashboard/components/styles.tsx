import { Tab, Tabs, styled } from "@mui/material";
import { colors } from "../../../style/colors";
import { CardStyleProps } from "./cards/cards";

export type CardSize = "s" | "m" | "l";

export const DefaultWrapper = styled("div", {
  shouldForwardProp: (prop) => prop !== "width" && prop !== "row",
})<{ width?: number; row?: boolean }>(({ theme, width, row }) => ({
  display: "inline-flex",
  flexDirection: row ? "row" : "column",
  borderRadius: "16px",
  background: theme.palette.mode === "light" ? colors.white : colors.gray700,
  padding: "16px",
  width: "max-content",
  maxWidth: width ? `${width}px` : "auto",
  height: "max-content",
  flexWrap: "wrap",
}));

export const DefaultContainer = styled("div", {
  shouldForwardProp: (prop) => prop !== "row",
})<{ row?: boolean }>(({ row }) => ({
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
  flexDirection: "column",
  borderRadius: "16px",
  border: `1px solid ${colors.gray200}`,
}));

export const DataWrapper = styled("div", {
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
    justifyContent: isColumn ? "center" : "space-between",
    alignItems: "center",
    flexDirection: isColumn ? "column" : "row",
    borderRadius: topRadius ? "16px 16px 0 0" : "16px",
    border: noBorder ? "none" : `1px solid ${colors.gray200}`,
    background:
      theme.palette.mode === "light"
        ? withoutBackground
          ? colors.white
          : colors.gray50
        : withoutBackground
        ? colors.gray700
        : colors.black,
    padding: "25px",
    minWidth: size === "l" ? "350px" : "210px",
    width: fullWidth ? "100%" : "auto",
    margin: withMargin ? "0 8px" : 0,

    "&:first-child": {
      margin: "0",
      marginRight: withMargin ? "8px" : 0,
    },
    "&:last-child": {
      margin: "0",
      marginLeft: withMargin ? "8px" : 0,
    },
  })
);

export const LayoutWrapper = styled("div")(() => ({
  display: "flex",
  flexDirection: "row",
  flexWrap: "wrap",
  padding: "25px",
  justifyContent: "center",
  width: "100%",

  "> div": {
    margin: "10px",
  },
}));

export const StyledTab = styled(Tab)(() => ({
  ".MuiButtonBase-root-MuiTab-root .Mui-selected": {
    backgroundColor: colors.gray50,
    borderRadius: "16px 16px 0px 0px",
    color: colors.batonGreen600,
  },
  ".Mui-selected": {
    backgroundColor: colors.gray50,
    borderRadius: "16px 16px 0px 0px",
    color: colors.batonGreen600,
  },
}));

export const StyledTabs = styled(Tabs)(() => ({
  ".MuiTabs-indicator": {
    display: "none",
  },
}));
