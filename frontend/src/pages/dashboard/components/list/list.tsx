import React from "react";

import Arrow from "@mui/icons-material/ArrowCircleRightOutlined";
import {
  ButtonWrapper,
  Count,
  ListItem,
  StyledList,
  StyledListItem,
  Title,
} from "./styles";
import { useLocation } from "react-router-dom";
import { normalizeString } from "../../../../common/helpers";

type ResourceListProps = {
  resources: any;
  resourceType?: string;
  remainingResources?: number;
};

type Suggestion = {
  text: string;
  link: string;
};

type SuggestionsListProps = {
  suggestions: Suggestion[];
};

export const ResourcesList = (props: ResourceListProps) => {
  const location = useLocation();
  return (
    <StyledList>
      {props.resourceType && (
        <Title>Users by {normalizeString(props.resourceType, false)}</Title>
      )}
      {props.resources.map((resource) => {
        return (
          <ListItem
            key={resource.resource.id.resource}
            to={`/${resource.resource_type.id}/${resource.resource.id.resource}`}
            state={{ from: location.pathname }}
          >
            <span>{resource.resource.display_name}</span>
            <Count>
              <span>{resource?.userCount || 0}</span>
              <Arrow fontSize="small" />
            </Count>
          </ListItem>
        );
      })}
      {props.remainingResources > 0 && (
        <StyledListItem
          to={`/${props.resourceType}`}
        >{`${props.remainingResources} more...`}</StyledListItem>
      )}
    </StyledList>
  );
};

export const SuggestionsList = (props: SuggestionsListProps) => {
  return (
    <StyledList>
      {props.suggestions.map((suggestion, i) => (
        <ListItem key={suggestion.text} to={`/${suggestion.link}`}>
          <span>{suggestion.text}</span>
          <ButtonWrapper>
            <Arrow fontSize="small" />
          </ButtonWrapper>
        </ListItem>
      ))}
    </StyledList>
  );
};
