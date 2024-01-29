import React, { useEffect, useState } from "react";
import List from "@mui/material/List";
import { Divider, Tooltip, useTheme } from "@mui/material";
import pluralize from "pluralize";
import {
  StyledDrawer,
  CloseButton,
  IconWrapper,
  NavWrapper,
  StyledLink,
} from "./styles";
import { ThemeSwitcher } from "./components/themeSwitcher";
import { Logo } from "./components/logo";
import { IconPerType } from "../icons/resourceTypeIcon";
import { DashboardButton } from "./components/dashboardButton";
import { useLocation } from "react-router-dom";

export const Navigation = ({ openResourceList, resourceState, closeResourceList }) => {
  const theme = useTheme();
  const [data, setData] = useState([]);
  const location = useLocation();
  const isDashboard = location.pathname === "/dashboard";

  useEffect(() => {
    const fetchData = async () => {
      const res = await (await fetch("/api/resourceTypes")).json();
      const resourceTypes = res.data.resource_types;
      const sorted = resourceTypes.sort((a, b) =>
        a.resource_type.display_name.localeCompare(b.resource_type.display_name)
      );
      setData(sorted);
    };
    fetchData();
  }, []);

    useEffect(() => {
      // Filter slashes so that trailing slashes behave the same as no trailing slash
      const splitPath = location.pathname.split("/").filter((x) => x !== "");
      const type = splitPath[0];
      if (type === resourceState.resource) {
        // Prevent infinite loop of opening/closing resource list
        return;
      }
      if (resourceState.opened) {
        closeResourceList()
      }
      const hasResourceId = splitPath.length > 1
      if (!isDashboard && !hasResourceId) {
        openResourceList(type);
      }
    }, [location.pathname, closeResourceList, openResourceList, isDashboard, resourceState.resource, resourceState.opened]);


  return (
    <StyledDrawer variant="permanent" theme={theme}>
      <NavWrapper>
        <Logo />
        <DashboardButton />
        <Divider flexItem variant="middle" />
        <List>
          {data?.map((type) => (
            <Tooltip
              key={type.resource_type.display_name}
              title={pluralize(type.resource_type.display_name)}
              placement="right"
            >
              <CloseButton
                onClick={() => openResourceList(type.resource_type.id)}
              >
                <StyledLink to={`/${type.resource_type.id}`}>
                  <IconWrapper
                    isSelected={
                      resourceState.resource === type.resource_type.id &&
                      !isDashboard
                    }
                  >
                    <IconPerType
                      color={
                        resourceState.resource === type.resource_type.id &&
                        !isDashboard
                          ? theme.palette.secondary.main
                          : theme.palette.mode === "light"
                          ? theme.palette.primary.dark
                          : theme.palette.primary.contrastText
                      }
                      resourceType={type.resource_type.id}
                      resourceTrait={
                        type.resource_type?.traits
                          ? type.resource_type?.traits[0]
                          : 0
                      }
                    />
                  </IconWrapper>
                </StyledLink>
              </CloseButton>
            </Tooltip>
          ))}
        </List>
      </NavWrapper>
      <ThemeSwitcher />
    </StyledDrawer>
  );
};
