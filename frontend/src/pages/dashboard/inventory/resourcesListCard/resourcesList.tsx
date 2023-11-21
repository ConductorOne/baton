import * as React from "react";
import { Card } from "../../components/cards/cards";
import { DefaultContainer, DefaultWrapper } from "../../components/styles";
import { ResourcesList } from "../../components/list/list";

export const ResourcesListCard = ({ resourceType, resourcesCount, resources }) => {
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
        />
        <ResourcesList resources={resources} count={"34"} />
      </DefaultContainer>
    </DefaultWrapper>
  );
};
