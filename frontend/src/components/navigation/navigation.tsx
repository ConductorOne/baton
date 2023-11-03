import React, { useEffect, useState } from "react";
import List from "@mui/material/List";
import { Tooltip, useTheme } from "@mui/material";
import pluralize from "pluralize";
import { StyledDrawer, CloseButton, IconWrapper, NavWrapper } from "./styles";
import { ThemeSwitcher } from "./components/themeSwitcher";
import { Logo } from "./components/logo";
import { IconPerType } from "../icons/resourceTypeIcon";

export const Navigation = ({ openResourceList, resourceState }) => {
  const theme = useTheme();
  const [data, setData] = useState([]);
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

  return (
    <StyledDrawer variant="permanent" theme={theme}>
      <NavWrapper>
        <Logo />
        <List>
          {data &&
            data.map((type) => (
              <Tooltip
                key={type.resource_type.display_name}
                title={pluralize(type.resource_type.display_name)}
                placement="right"
              >
                <CloseButton
                  onClick={() => openResourceList(type.resource_type.id)}
                >
                  <IconWrapper
                    isSelected={
                      resourceState.resource === type.resource_type.id
                    }
                  >
                    <IconPerType
                      color={
                        resourceState.resource === type.resource_type.id
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
                </CloseButton>
              </Tooltip>
            ))}
        </List>
      </NavWrapper>
      <ThemeSwitcher />
    </StyledDrawer>
  );
};
