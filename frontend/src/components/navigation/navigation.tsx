import React, { useEffect, useState } from "react";
import List from "@mui/material/List";
import { BatonLogo, ResourceTypeIcon } from "../icons/icons";
import { Tooltip, useTheme } from "@mui/material";
import pluralize from "pluralize";
import { Logo, StyledDrawer, CloseButton } from "./styles";
import { IconWrapper } from "../explorer/styles/styles";

export const Navigation = ({ openResourceList }) => {
  const theme = useTheme();
  const [data, setData] = useState([]);
  useEffect(() => {
    const fetchData = async () => {
      const res = await (await fetch("/api/resourceTypes")).json();
      const resourceTypes = res.data.resource_types
      const sorted = resourceTypes.sort((a, b) =>
        a.resource_type.display_name.localeCompare(b.resource_type.display_name)
      );
      setData(sorted);
    };
    fetchData();
  }, []);

  return (
    <StyledDrawer variant="permanent" theme={theme}>
      <Logo>
        <BatonLogo color="primary.contrastText" />
      </Logo>
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
                <IconWrapper sx={{marginRight: 0}}>
                <ResourceTypeIcon
                  size="medium"
                  resourceTrait={
                    type.resource_type?.traits
                      ? type.resource_type?.traits[0]
                      : 0
                  }
                  color="icon"
                />
                </IconWrapper>
              </CloseButton>
            </Tooltip>
          ))}
      </List>
    </StyledDrawer>
  );
};
