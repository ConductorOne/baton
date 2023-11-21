import React, { useContext, useState } from "react";
import { ExplorerPage } from "./pages/explorerPage";
import { themeModePalette } from "./style/theme";
import { ThemeProvider } from "@mui/material/styles";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import { ResourcesContextProvider } from "./context/resources";
import { CssBaseline, createTheme } from "@mui/material";
import { Dashboard } from "./pages/dashboard/dashboard";
import { Navigation } from "./components/navigation/navigation";
import Footer from "./components/footer/footer";

const ColorModeContext = React.createContext({ toggleColorMode: () => {} });
type ResourcesListState = {
  opened: boolean;
  resource?: string;
};

function App() {
  const [mode, setMode] = useState("light");
  const [resourceList, setResourceList] = useState<ResourcesListState>({
    opened: false,
  });

  const colorMode = React.useMemo(
    () => ({
      toggleColorMode: () => {
        setMode((prevMode) => (prevMode === "light" ? "dark" : "light"));
      },
    }),
    []
  );

  const theme = React.useMemo(
    () => createTheme(themeModePalette(mode)),
    [mode]
  );

    const openResourceList = (resourceType: string) => {
      setResourceList({
        opened: true,
        resource: resourceType,
      });
    };

    const closeResourceList = () => {
      setResourceList({
        opened: false,
      });
    };

  return (
    <ColorModeContext.Provider value={colorMode}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <ResourcesContextProvider>
          <div className="App">
            <Router>
              <Navigation
                openResourceList={openResourceList}
                resourceState={resourceList}
              />
              <Routes>
                <Route path="/dashboard" Component={Dashboard} />
                <Route
                  path="/:type?/:id?"
                  element={
                    <ExplorerPage
                      resourceList={resourceList}
                      closeResourceList={closeResourceList}
                    />
                  }
                />
              </Routes>
              <Footer />
            </Router>
          </div>
        </ResourcesContextProvider>
      </ThemeProvider>
    </ColorModeContext.Provider>
  );
}

export const useColorModeContext = () => useContext(ColorModeContext);
export default App;
