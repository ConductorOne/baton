import * as React from "react";
import { PieTypeChart } from "../../components/charts/pieChart";
import { DefaultContainer, DefaultWrapper } from "../../components/styles";
import { ChartWrapper } from "../../inventory/resourcesGraph/styles";
import { Data } from "./style";
import { ResourcesList } from "../../components/list/list";
import { colors } from "../../../../style/colors";

type GraphCardProps = {
  header: string;
  count: number;
  users: any[];
  chartData: any;
  text: string | number;
};

export const CardWithGraph = (props: GraphCardProps) => {
  return (
    <DefaultWrapper width={450}>
      <DefaultContainer>
        <ChartWrapper>
          <PieTypeChart
            data={props.chartData}
            colors={[
              colors.orange400,
              colors.orange500,
              colors.orange600,
              colors.orange700,
            ]}
            width={350}
            height={350}
            textFillColor={colors.orange600}
            textPositionX={175}
            textPositionY={187}
            textSize={48}
            text={props.text}
          />
        </ChartWrapper>
        <>
          <Data>
            total {props.header}
            <span>{props.count}</span>
          </Data>
          <p style={{ display: "flex", alignSelf: "flex-start", padding: "0 16px"}}>USERS</p>
          <ResourcesList resources={props.users} />
        </>
      </DefaultContainer>
    </DefaultWrapper>
  );
};
