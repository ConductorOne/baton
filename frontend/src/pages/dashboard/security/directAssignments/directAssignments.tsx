import * as React from "react";
import { DefaultContainer, DefaultWrapper } from "../../components/styles";
import { useTheme } from "@mui/material";
import { ChartWrapper } from "../../inventory/resourcesGraph/styles";
import { BarTypeChart } from "../../components/charts/barChart";
import { ResourcesTable } from "./table";

const data = [
  {
    name: "Repositories",
    allResources: 250,
    directAssigments: 32,
  },
  {
    name: "Projects",
    allResources: 170,
    directAssigments: 101,
  },
  {
    name: "Vaults",
    allResources: 105,
    directAssigments: 23,
  },
  {
    name: "Websites",
    allResources: 120,
    directAssigments: 79,
  },
  {
    name: "IDPs",
    allResources:115,
    directAssigments: 5,
  },
];


export const DirectAssignments = () => {
  const theme = useTheme();
  const lightMode = theme.palette.mode === "light";
  return (
    <DefaultWrapper width={790}>
      <DefaultContainer>
        <ChartWrapper>
          <BarTypeChart data={data} width={650} height={270}/>
        </ChartWrapper>
        <ResourcesTable />
      </DefaultContainer>
    </DefaultWrapper>
  );
};
