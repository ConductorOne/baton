import { colors } from "./colors";

export const themeModePalette = (mode) => ({
  palette: {
    mode,
    ...(mode === 'light'
      ? {
        background: {
          default: colors.gray100,
        },
        primary: {
          main: colors.white,
          dark: colors.gray500,
          contrastText: colors.gray900,
        },
        secondary: {
          main: colors.batonGreen600,
          dark: colors.batonGreen700,
          light: colors.batonGreen100,
          contrastText: colors.white,
        },
      }
      : {
        background: {
          default: colors.gray950,
        },
        primary: {
          main: colors.gray800,
          dark: colors.black,
          contrastText: colors.white
        },
        secondary: {
          main: colors.batonGreen600,
          dark: colors.batonGreen700,
          light: colors.batonGreen100,
          contrastText: colors.white,
        },
      }
    ),
  },
typography: {
  h5: {
    fontSize: '18px',
    fontWeight: 500,
  },
  },
  components: {
    MuiMenu: {
      styleOverrides: {
        paper: {
          borderRadius: "8px",
          backgroundColor: colors.batonGreen600,
          boxShadow: "0px 6px 8px -4px rgba(16, 24, 40, 0.03), 0px 10px 24px -4px rgba(16, 24, 40, 0.19)",
          color: "white",
          marginTop: "5px",
        }
      },
    },
    MuiSelect: {
      styleOverrides: {
        root: {
          fontSize: "12px",
          lineHeight: "inherit",
          padding: "0 10px",
          border: `1px solid ${colors.batonGreen600}`,
          borderRadius: "100px",
          display: "block",
          backgroundColor: mode === "light" ? colors.white : colors.black,
        },
        icon: {
          display: "none"
        },
        select: {
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          paddingRight: "0px !important",
          ":focus": {
            backgroundColor: "transparent"
          }
        },
        nativeInput: {
          paddingRight: "0",
        }
      }
    },
    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          backgroundColor: colors.batonGreen600,
          color: colors.white,
        }
      }
    },
    MuiIconButton: {
      styleOverrides: {
        root: {
          border: `1px solid ${colors.gray300}`
        },
      }
    },
    MuiTab: {
      styleOverrides: {
        root: {
          padding: "0 16px",
          minHeight: "60px",
          "&.Mui-selected": {
            background: colors.gray50,
            borderRadius: "16px 16px 0 0",
            color: colors.batonGreen600,
            svg: {
              fill: colors.white,
            }
          },
        },
      }
    },
  },
});
