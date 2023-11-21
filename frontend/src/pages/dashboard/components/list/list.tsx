import React from "react";

import Arrow from "@mui/icons-material/ArrowCircleRightOutlined";
import { ButtonWrapper, Count, ListItem, StyledList } from "./styles";

type ResourcesListProps = {
  resources: any,
  count?: any
}

type Suggestion = {
  text: string
  link: string
}

type SuggestionsListProps = {
  suggestions: Suggestion[],
};

export const ResourcesList = (props: ResourcesListProps) => {
  const { resources, count } = props;
  return (
    <StyledList>
      {resources.map(resource => (
        <ListItem
          key={resource.resource.id.resource}
          to={`/${resource.resource_type}/${resource.resource.id.resource}`}
        >
          <span>{resource.resource.display_name}</span>
          <Count>
            <span>{count}</span>
            <Arrow fontSize="small" />
          </Count>
        </ListItem>
      ))}
    </StyledList>
  );
};

export const SuggestionsList = (props: SuggestionsListProps) => {
  return (
    <StyledList>
      {props.suggestions.map((suggestion, i) => (
        <ListItem
          key={i}
          to={`/${suggestion.link}`}
        >
          <span>{suggestion.text}</span>
          <ButtonWrapper>
            <Arrow fontSize="small" />
          </ButtonWrapper>
        </ListItem>
      ))}
    </StyledList>
  );
};