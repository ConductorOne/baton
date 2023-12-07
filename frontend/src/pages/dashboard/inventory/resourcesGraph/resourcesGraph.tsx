import * as React from "react";
import { ChartWrapper, DataWrapper } from "./styles";
import { PieTypeChart } from "../../components/charts/pieChart";
import { Card } from "../../components/cards/cards";
import { DefaultContainer, DefaultWrapper } from "../../components/styles";
import { colors } from "../../../../style/colors";
import { useTheme } from "@mui/material";
import { useResources } from "../../../../context/resources";
import pluralize from "pluralize";
import { normalizeString } from "../../../../common/helpers";
import { Label } from "../../components/cards/styles";

const graphColors = [
  colors.batonGreen600,
  colors.batonGreen800,
  colors.pink300,
  colors.blue500,
  colors.purple400,
  colors.blue200,
  colors.cyan400,
  colors.indigo500,
];

export const Resources = () => {
  const theme = useTheme();
  const lightMode = theme.palette.mode === "light";
  const { resources, mappedResources } = useResources();
  let graphData = [];
  Object.keys(mappedResources).forEach((type) =>
    graphData.push({
      name: pluralize(normalizeString(type, false)),
      value: mappedResources[type].length,
      type: type,
    })
  );

  return (
    <DefaultWrapper width={600}>
      <DefaultContainer>
        <ChartWrapper>
          <Label
            size={{
              fontSize: "20px",
            }}
          >
            Resources
          </Label>
          <PieTypeChart
            data={graphData && graphData}
            colors={graphColors}
            width={568}
            height={250}
            textPositionX={283}
            textPositionY={140}
            textFillColor={
              lightMode ? colors.batonGreen900 : colors.batonGreen500
            }
            textSize={48}
            text={resources?.resources?.length}
          />
        </ChartWrapper>
        <DataWrapper>
          {graphData.map((d, i) => (
            <Card
              to={`/${d.type}`}
              key={d.name}
              isColumn
              label={d.name}
              count={d.value}
              size="s"
              noBorder
              withoutBackground
            />
          ))}
        </DataWrapper>
      </DefaultContainer>
    </DefaultWrapper>
  );
};
