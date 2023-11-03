import React, { useContext } from "react";
import Homepage from "./pages/homepage";
import { themeModePalette } from "./style/theme";
import { ThemeProvider } from "@mui/material/styles";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import { ResourcesContextProvider } from "./context/resources";
import { CssBaseline, createTheme } from "@mui/material";

const ColorModeContext = React.createContext({ toggleColorMode: () => {} });

function App() {
  const [mode, setMode] = React.useState("light");
  const colorMode = React.useMemo(
    () => ({
      toggleColorMode: () => {
        setMode((prevMode) =>
          prevMode === "light" ? "dark" : "light"
        );
      },
    }),
    []
  );

  const theme = React.useMemo(
    () => createTheme(themeModePalette(mode)),
    [mode]
  );

  return (
    <ColorModeContext.Provider value={colorMode}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <ResourcesContextProvider>
          <div className="App">
            <Router>
              <Routes>
                <Route path="/:type?/:id?" Component={Homepage} />
              </Routes>
            </Router>
          </div>
        </ResourcesContextProvider>
      </ThemeProvider>
    </ColorModeContext.Provider>
  );
}

export const useColorModeContext = () => useContext(ColorModeContext);
export default App;
