import { Typography, styled } from "@mui/material";
import React, { useState } from "react";
import { Inventory } from "./inventory/inventory";
import { colors } from "../../style/colors";
import Rocket from "@mui/icons-material/RocketLaunchOutlined";
import People from "@mui/icons-material/PeopleOutlineOutlined";
import Lock from "@mui/icons-material/LockOutlined";
import Shield from "@mui/icons-material/ShieldOutlined";
import ServiceAccount from "@mui/icons-material/ManageAccountsOutlined";
import { DashboardTabs } from "./components/tabs/tabs";
import { Security } from "./security/security";
import { NonHumanIdentity } from "./nonHumanIdentity/nonHumanIdentity";

export const Wrapper = styled("div")(({ theme }) => ({
  display: "flex",
  height: "100%",
  marginLeft: "78px",
  flexDirection: "column",
  background: theme.palette.mode === "light" ? colors.white : colors.gray800,

  h6: {
    fontSize: "20px",
    fontWeight: 700,
    marginBottom: "20px",
  },
}));

export const Content = styled("div")(({ theme }) => ({
  display: "flex",
  height: "100%",
  background: theme.palette.mode === "light" ? colors.gray100 : colors.gray950,
}));

export const Header = styled("div")(() => ({
  display: "flex",
  height: "100%",
  flexDirection: "column",
  padding: "20px 0 0 20px",
}));

const tabs = {
  inventory: {
    label: "Inventory",
    icon: <People />,
    value: "inventory",
    tabPanel: <Inventory />,
  },
  security: {
    label: "Security",
    icon: <Lock />,
    value: "security",
    tabPanel: <Security />,
  },
  nhidentities: {
    label: "Non-Human Identity",
    icon: <ServiceAccount />,
    value: "nhidentities",
    tabPanel: <NonHumanIdentity />
  },
  // apps: {
  //   label: "Applications",
  //   icon: <Rocket />,
  //   value: "apps",
  //   tabPanel: <div>Placeholder</div>,
  // },
  // risks: {
  //   label: "Risk & Threats",
  //   icon: <Shield />,
  //   value: "risks",
  //   tabPanel: <div>Placeholder</div>,
  // },
};

export const Dashboard = () => {
  const [selectedTab, setSelectedTab] = useState("inventory")

  const handleChange = (e, tab) => {
    setSelectedTab(tab);
  };

  return (
    <Wrapper>
      <Header>
        <Typography variant="h6">Dashboard</Typography>
        <DashboardTabs tabs={tabs} value={selectedTab} handleChange={handleChange} />
      </Header>
      <Content>{tabs[selectedTab].tabPanel}</Content>
    </Wrapper>
  );
};
