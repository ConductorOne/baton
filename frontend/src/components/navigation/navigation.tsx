import React, { useEffect, useState } from "react";
import List from "@mui/material/List";
import { BatonLogo, ResourceTypeIcon } from "../icons/icons";
import { Tooltip, useTheme } from "@mui/material";
import pluralize from "pluralize";
import { Logo, StyledDrawer, CloseButton } from "./styles";

interface ResourceTypesData {
  resource_types: {
    resource_type: {
      id: string;
      display_name: string;
      traits?: number[];
    };
  }[];
}

export const Navigation = ({ openResourceList }) => {
  const theme = useTheme();
  const [data, setData] = useState<ResourceTypesData>();
  useEffect(() => {
    const fetchData = async () => {
      const res = await (await fetch("/api/resourceTypes")).json();
      setData(res.data);
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
          data.resource_types.map((type) => (
            <Tooltip
              key={type.resource_type.display_name}
              title={pluralize(type.resource_type.display_name)}
              placement="right"
            >
              <CloseButton
                onClick={() => openResourceList(type.resource_type.id)}
              >
                <ResourceTypeIcon
                  size="medium"
                  resourceTrait={
                    type.resource_type?.traits
                      ? type.resource_type?.traits[0]
                      : 0
                  }
                  color="icon"
                />
              </CloseButton>
            </Tooltip>
          ))}
      </List>
    </StyledDrawer>
  );
};
