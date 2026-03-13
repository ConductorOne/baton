import { colors } from "./colors";

export const themeModePalette = (mode) => ({
  palette: {
    mode,
    ...(mode === "light"
      ? {
          background: {
            default: colors.white,
            paper: colors.white,
          },
          primary: {
            main: colors.white,
            dark: colors.gray600,
            contrastText: colors.gray900,
          },
          secondary: {
            main: colors.batonGreen600,
            dark: colors.batonGreen700,
            light: colors.batonGreen100,
            contrastText: colors.white,
          },
          text: {
            primary: colors.gray900,
            secondary: colors.gray500,
          },
          divider: colors.gray200,
        }
      : {
          background: {
            default: colors.gray900,
            paper: colors.gray800,
          },
          primary: {
            main: colors.gray800,
            dark: colors.gray950,
            contrastText: colors.gray50,
          },
          secondary: {
            main: colors.batonGreen500,
            dark: colors.batonGreen600,
            light: colors.batonGreen200,
            contrastText: colors.gray900,
          },
          text: {
            primary: colors.gray50,
            secondary: colors.gray400,
          },
          divider: "rgba(255,255,255,0.08)",
        }),
  },
  typography: {
    fontFamily: '"Inter", "Helvetica", "Arial", sans-serif',
    h5: {
      fontSize: "18px",
      fontWeight: 600,
    },
    body2: {
      fontSize: "13px",
    },
  },
  shape: {
    borderRadius: 8,
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          backgroundColor: mode === "light" ? colors.white : colors.gray900,
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        outlined: {
          borderColor: mode === "light" ? colors.gray200 : "rgba(255,255,255,0.08)",
        },
      },
    },
    MuiTableHead: {
      styleOverrides: {
        root: {
          "& .MuiTableCell-head": {
            backgroundColor: mode === "light" ? colors.gray50 : colors.gray800,
            color: mode === "light" ? colors.gray600 : colors.gray300,
            fontWeight: 600,
            fontSize: "12px",
            textTransform: "uppercase" as const,
            letterSpacing: "0.5px",
            borderBottom: `1px solid ${mode === "light" ? colors.gray200 : "rgba(255,255,255,0.08)"}`,
            padding: "10px 16px",
          },
        },
      },
    },
    MuiTableBody: {
      styleOverrides: {
        root: {
          "& .MuiTableRow-root": {
            borderBottom: `1px solid ${mode === "light" ? colors.gray100 : "rgba(255,255,255,0.04)"}`,
            "&:nth-of-type(even)": {
              backgroundColor: mode === "light" ? colors.gray25 : "rgba(255,255,255,0.02)",
            },
            "&:hover": {
              backgroundColor: mode === "light" ? colors.batonGreen100 : "rgba(155,237,117,0.06)",
            },
          },
          "& .MuiTableCell-body": {
            borderBottom: "none",
            padding: "8px 16px",
            fontSize: "13px",
          },
        },
      },
    },
    MuiTableSortLabel: {
      styleOverrides: {
        root: {
          color: mode === "light" ? colors.gray600 : colors.gray300,
          "&.Mui-active": {
            color: mode === "light" ? colors.gray900 : colors.gray50,
          },
        },
      },
    },
    MuiMenu: {
      styleOverrides: {
        paper: {
          borderRadius: "8px",
          backgroundColor: mode === "light" ? colors.white : colors.gray700,
          boxShadow:
            "0px 4px 16px rgba(0,0,0,0.12), 0px 1px 4px rgba(0,0,0,0.08)",
          border: `1px solid ${mode === "light" ? colors.gray200 : "rgba(255,255,255,0.08)"}`,
          color: mode === "light" ? colors.gray900 : colors.gray50,
          marginTop: "4px",
        },
      },
    },
    MuiMenuItem: {
      styleOverrides: {
        root: {
          fontSize: "13px",
          "&:hover": {
            backgroundColor: mode === "light" ? colors.gray50 : "rgba(255,255,255,0.06)",
          },
          "&.Mui-selected": {
            backgroundColor: mode === "light" ? colors.batonGreen100 : "rgba(155,237,117,0.12)",
            "&:hover": {
              backgroundColor: mode === "light" ? colors.batonGreen200 : "rgba(155,237,117,0.16)",
            },
          },
        },
      },
    },
    MuiSelect: {
      styleOverrides: {
        root: {
          fontSize: "13px",
          lineHeight: "inherit",
          border: `1px solid ${mode === "light" ? colors.gray200 : "rgba(255,255,255,0.12)"}`,
          borderRadius: "8px",
          backgroundColor: mode === "light" ? colors.white : colors.gray800,
          "&:hover": {
            borderColor: mode === "light" ? colors.gray400 : "rgba(255,255,255,0.2)",
          },
        },
        icon: {
          color: mode === "light" ? colors.gray400 : colors.gray500,
        },
        select: {
          display: "flex",
          alignItems: "center",
          ":focus": {
            backgroundColor: "transparent",
          },
        },
      },
    },
    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          backgroundColor: mode === "light" ? colors.gray900 : colors.gray700,
          color: colors.white,
          fontSize: "12px",
          borderRadius: "6px",
          padding: "6px 10px",
        },
      },
    },
    MuiIconButton: {
      styleOverrides: {
        root: {
          border: "none",
          borderRadius: "8px",
          "&:hover": {
            backgroundColor: mode === "light" ? colors.gray100 : "rgba(255,255,255,0.06)",
          },
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          fontWeight: 500,
        },
        outlined: {
          borderColor: mode === "light" ? colors.gray200 : "rgba(255,255,255,0.12)",
        },
      },
    },
    MuiTab: {
      styleOverrides: {
        root: {
          padding: "0 16px",
          minHeight: "48px",
          color: mode === "light" ? colors.gray500 : colors.gray400,
          fontWeight: 500,
          textTransform: "none" as const,
          "&.Mui-selected": {
            color: mode === "light" ? colors.batonGreen700 : colors.batonGreen500,
            fontWeight: 600,
          },
        },
      },
    },
    MuiToggleButton: {
      styleOverrides: {
        root: {
          border: `1px solid ${mode === "light" ? colors.gray200 : "rgba(255,255,255,0.12)"}`,
          color: mode === "light" ? colors.gray500 : colors.gray400,
          textTransform: "none" as const,
          padding: "4px 10px",
          "&.Mui-selected": {
            backgroundColor: mode === "light" ? colors.batonGreen100 : "rgba(155,237,117,0.12)",
            color: mode === "light" ? colors.batonGreen700 : colors.batonGreen500,
            borderColor: mode === "light" ? colors.batonGreen300 : colors.batonGreen600,
            "&:hover": {
              backgroundColor: mode === "light" ? colors.batonGreen200 : "rgba(155,237,117,0.16)",
            },
          },
          "&:hover": {
            backgroundColor: mode === "light" ? colors.gray50 : "rgba(255,255,255,0.04)",
          },
          "&.Mui-disabled": {
            color: mode === "light" ? colors.gray300 : colors.gray600,
            borderColor: mode === "light" ? colors.gray100 : "rgba(255,255,255,0.04)",
          },
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          "& .MuiOutlinedInput-root": {
            borderRadius: "8px",
            fontSize: "13px",
            "& fieldset": {
              borderColor: mode === "light" ? colors.gray200 : "rgba(255,255,255,0.12)",
            },
            "&:hover fieldset": {
              borderColor: mode === "light" ? colors.gray400 : "rgba(255,255,255,0.2)",
            },
            "&.Mui-focused fieldset": {
              borderColor: mode === "light" ? colors.batonGreen600 : colors.batonGreen500,
              borderWidth: "1px",
            },
          },
        },
      },
    },
    MuiDivider: {
      styleOverrides: {
        root: {
          borderColor: mode === "light" ? colors.gray200 : "rgba(255,255,255,0.08)",
        },
      },
    },
  },
});
