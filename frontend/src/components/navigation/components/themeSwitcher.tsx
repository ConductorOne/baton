import React from "react";
import { colors } from "../../../style/colors"

import {
  Switch,
  styled,
  useTheme,
} from "@mui/material";
import { useColorModeContext } from "../../../App";

const CustomSwitch = styled(Switch)(({ theme }) => ({
  padding: "8px",
  "& .MuiSwitch-track": {
    borderRadius: "12px",
    backgroundColor:
      theme.palette.mode === "light"
        ? `${colors.gray200} !important`
        : `${colors.gray950} !important`,
    "&:before, &:after": {
      content: '""',
      position: "absolute",
      top: "50%",
      transform: "translateY(-50%)",
      width: "12px",
      height: "12px",
      backgroundRepeat: "no-repeat",
    },
    "&:before": {
      backgroundImage: `url('data:image/svg+xml,<svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg"><g clip-path="url(%23clip0_7545_1294)"><path d="M6 1V2M6 10V11M2 6H1M3.15706 3.15706L2.44995 2.44995M8.84294 3.15706L9.55005 2.44995M3.15706 8.845L2.44995 9.55211M8.84294 8.845L9.55005 9.55211M11 6H10M8.5 6C8.5 7.38071 7.38071 8.5 6 8.5C4.61929 8.5 3.5 7.38071 3.5 6C3.5 4.61929 4.61929 3.5 6 3.5C7.38071 3.5 8.5 4.61929 8.5 6Z" stroke="%23FFFFFF" stroke-linecap="round" stroke-linejoin="round" /></g><defs><clipPath id="clip0_7545_1294"><rect width="12" height="12" fill="white" /></clipPath></defs></svg>')`,
      left: "12px",
      svg: {
        fill: colors.gray25,
      },
    },
    "&:after": {
      backgroundImage: `url('data:image/svg+xml,<svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M10.9774 6.47831C10.289 7.68595 8.98957 8.50016 7.5 8.50016C5.29086 8.50016 3.5 6.7093 3.5 4.50016C3.5 3.01048 4.31434 1.711 5.52213 1.02258C2.98487 1.26316 1 3.3998 1 6.00004C1 8.76147 3.23858 11 6 11C8.60011 11 10.7367 9.01538 10.9774 6.47831Z" stroke="%23101828" stroke-linecap="round" stroke-linejoin="round"/></svg>')`,
      right: "12px",
      svg: {
        fill: colors.gray900,
      }
    },
  },
  "& .MuiSwitch-thumb": {
    backgroundColor:
      theme.palette.mode === "light" ? colors.white : colors.gray300,
    width: "16px",
    height: "16px",
    margin: "2px",
  },
  "& .MuiSwitch-switchBase": {
    "&:hover": {
      backgroundColor: "transparent",
    },
  },
}));

export const ThemeSwitcher = () => {
  const theme = useTheme();
  const colorMode = useColorModeContext();

  return (
    <CustomSwitch
      onChange={colorMode.toggleColorMode}
      value={theme.palette.mode}
    />
  );
};
