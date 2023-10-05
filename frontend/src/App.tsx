import React from "react";
import Homepage from "./pages/homepage";
import theme from "./styles/theme";
import { ThemeProvider } from "@mui/material/styles";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import { ResourcesContextProvider } from "./context/resources";

function App() {
  return (
    <ThemeProvider theme={theme}>
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
  );
}

export default App;
