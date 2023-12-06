import React from "react";
import { StyledTabs } from "../styles";
import { IconWrapper } from "../../../../components/explorer/styles/styles";
import { colors } from "../../../../style/colors";
import { Tab, useTheme } from "@mui/material";

export const DashboardTabs = ({ tabs, value, handleChange }) => {
  const theme = useTheme();
  return (
    <StyledTabs value={value} onChange={handleChange} textColor="inherit">
      {Object.keys(tabs).map((tab) => {
        const defaultBgColor =
          theme.palette.mode === "light" ? colors.gray100 : colors.black;
        const tabSelected = value === tabs[tab].value;
        return (
          <Tab
            key={tabs[tab].value}
            icon={
              <IconWrapper
                backgroundColor={
                  tabSelected ? colors.batonGreen600 : defaultBgColor
                }
              >
                {tabs[tab].icon}
              </IconWrapper>
            }
            iconPosition="start"
            value={tabs[tab].value}
            label={tabs[tab].label}
          />
        );
      })}
    </StyledTabs>
  );
};
