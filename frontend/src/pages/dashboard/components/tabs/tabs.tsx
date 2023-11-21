import React from "react";
import { StyledTab, StyledTabs } from "../styles";
import { IconWrapper } from "../../../../components/explorer/styles/styles";
import { colors } from "../../../../style/colors";

export const DashboardTabs = ({ tabs, value, handleChange }) => {
  return (
    <StyledTabs
        value={value}
        onChange={handleChange}
        textColor="inherit"
      >
        {Object.keys(tabs).map((tab) => {
          return (
            <StyledTab
              key={tabs[tab].value}
              icon={<IconWrapper backgroundColor={value === tabs[tab].value ? colors.batonGreen600 : colors.gray100}>{tabs[tab].icon}</IconWrapper>}
              iconPosition="start"
              value={tabs[tab].value}
              label={tabs[tab].label}
            />
          );
        })}
      </StyledTabs>
  );
};


