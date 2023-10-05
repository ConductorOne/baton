import { createTheme } from '@mui/material/styles';

const theme = createTheme({
  palette: {
    primary: {
      main: 'rgba(115, 189, 81, 1)',
      dark: 'rgba(98, 162, 68, 1)',
      light: 'rgba(228, 244, 221, 1)',
      contrastText: '#FFFFFF'
    },
    secondary: {
      main: '#475467',
      dark: '#101828'
    },
    icon: {
      main: 'rgba(255,255,255, 0.8)',
      dark: 'rgba(0, 0, 0, 0.60)',
    }
  },
  components: {
    MuiMenu: {
      styleOverrides: {
        paper: {
          borderRadius: "8px",
          backgroundColor: "rgba(115, 189, 81, 1)",
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
          border: "1px solid rgba(115, 189, 81, 1)",
          borderRadius: "100px",
          display: "block",
          backgroundColor: "rgba(255, 255, 255, 1)",
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
          backgroundColor: "rgba(115, 189, 81, 1)",
          color: "rgba(255, 255, 255, 1)",
        }
      }
    }
  },
});

export default theme;