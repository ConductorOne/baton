import React from "react";
import { CardSize, DataWrapper } from "../styles";
import { Count, Label, Score, SizeMap } from "./styles";
import { normalizeString } from "../../../../common/helpers";
import pluralize from "pluralize";

export type CardProps = {
  label: any;
  count: any;
  isScore?: boolean;
  to?: string;
} & CardStyleProps;

export type CardStyleProps = {
  size: CardSize;
  isColumn?: boolean;
  noBorder?: boolean;
  topRadius?: boolean;
  withoutBackground?: boolean;
  fullWidth?: boolean;
  withMargin?: boolean;
};

export const Card = (props: CardProps) => {
  return (
    <DataWrapper
      to={props.to}
      state={{ resourceType: props.label }}
      isColumn={props.isColumn}
      size={props.size}
      withoutBackground={props.withoutBackground}
      noBorder={props.noBorder}
      topRadius={props.topRadius}
      fullWidth={props.fullWidth}
      withMargin={props.withMargin}
    >
      <Label size={SizeMap[props.size].label}>
        {pluralize(normalizeString(props.label, false))}
      </Label>
      {props.isScore ? (
        <Score>
          <span>{props.count}</span> / 100
        </Score>
      ) : (
        <Count size={SizeMap[props.size].count}>{props.count}</Count>
      )}
    </DataWrapper>
  );
};
