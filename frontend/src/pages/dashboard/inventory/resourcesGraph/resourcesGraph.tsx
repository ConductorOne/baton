import * as React from "react";
import { ChartWrapper, DataWrapper } from "./styles";
import { PieTypeChart } from "../../components/charts/pieChart";
import { Card } from "../../components/cards/cards";
import { DefaultContainer, DefaultWrapper } from "../../components/styles";
import { colors } from "../../../../style/colors";
import { useTheme } from "@mui/material";

const data = [
  { name: "repositories", value: 63 },
  { name: "projects", value: 105 },
  { name: "vaults", value: 2 },
  { name: "websites", value: 63 },
  { name: "idps", value: 105 },
];
const graphColors = [
  colors.batonGreen600,
  colors.batonGreen800,
  colors.pink300,
  colors.blue500,
  colors.purple400,
  colors.blue200,
  colors.cyan400,
  colors.indigo500
];

export const Resources = () => {
  const theme = useTheme()
  const lightMode = theme.palette.mode === "light"
  
  return (
    <DefaultWrapper width={600}>
      <DefaultContainer>
        <ChartWrapper>
          <PieTypeChart
            data={data}
            colors={graphColors}
            width={568}
            height={290}
            textPositionX={283}
            textPositionY={160}
            textFillColor={
              lightMode ? colors.batonGreen900 : colors.batonGreen500
            }
            textSize={48}
            text={354}
          />
        </ChartWrapper>
        <DataWrapper>
          <Card
            isColumn
            label="repositories"
            count="63"
            size="s"
            noBorder
            withoutBackground
          />
          <Card
            isColumn
            label="projects"
            count="34"
            size="s"
            noBorder
            withoutBackground
          />
          <Card
            isColumn
            label="vaults"
            count="563"
            size="s"
            noBorder
            withoutBackground
          />
          <Card
            isColumn
            label="websites"
            count="643"
            size="s"
            noBorder
            withoutBackground
          />
          <Card
            isColumn
            label="idps"
            count="653"
            size="s"
            noBorder
            withoutBackground
          />
        </DataWrapper>
      </DefaultContainer>
    </DefaultWrapper>
  );
};
