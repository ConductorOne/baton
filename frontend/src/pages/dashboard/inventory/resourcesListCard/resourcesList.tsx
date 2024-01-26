import * as React from "react";
import { Card } from "../../components/cards/cards";
import { DefaultContainer, DefaultWrapper } from "../../components/styles";
import { ResourcesList } from "../../components/list/list";

type ResourcesListProps = {
  resourceType: string;
  resourcesCount: any;
  resources: any;
};

export const ResourcesListCard = (props: ResourcesListProps) => {
  const { resourceType, resourcesCount, resources } = props;
  const maxResources = 10;
  const sorted = resources.sort((a, b) => b.userCount - a.userCount);
  const cappedResources =
    sorted.length > maxResources ? sorted.slice(0, maxResources) : sorted;
  const remainingResources =
    sorted.length > maxResources ? sorted.length - maxResources : 0;

  return (
    <DefaultWrapper width={330}>
      <DefaultContainer>
        <Card
          isColumn
          topRadius
          size="m"
          label={resourceType}
          count={resourcesCount}
          fullWidth
          to={`/${resourceType}`}
        />
        <ResourcesList
          resources={cappedResources}
          remainingResources={remainingResources}
          resourceType={resourceType}
        />
      </DefaultContainer>
    </DefaultWrapper>
  );
};
