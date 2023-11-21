import React from "react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
} from "recharts";
import { colors } from "../../../../style/colors";

type BarChartProps = {
  data: any;
  width: number;
  height: number;
  colors?: string[];
};

export const BarTypeChart = (props: BarChartProps) => {
  return (
    <BarChart
      width={props.width}
      height={props.height}
      data={props.data}
      layout="vertical"
      barCategoryGap={"2%"}
      margin={{
        right: 10,
        top: 10,
        left: 10,
      }}
    >
      <XAxis type="number" axisLine={false} tickLine={false} />
      <YAxis
        tickLine={false}
        dataKey="name"
        type="category"
        padding={{ top: 10, bottom: 10 }}
        width={100}
        axisLine={false}
      />
      <Tooltip cursor={false} />
      <CartesianGrid
        horizontal={false}
        stroke={colors.gray200}
        fillOpacity={0.5}
      />
      <Bar
        activeBar={false}
        dataKey="directAssigments"
        fill={colors.orange600}
        stackId="a"
        barSize={20}
        fillOpacity={1}
      />
      <Bar
        dataKey="allResources"
        activeBar={false}
        fill={colors.gray200}
        stackId="a"
        barSize={20}
        fillOpacity={1}
      />
    </BarChart>
  );
};
