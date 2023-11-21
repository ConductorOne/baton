import * as React from "react";
import { DefaultContainer, DefaultWrapper } from "../components/styles";
import { Card } from "../components/cards/cards";
import { SuggestionsList } from "../components/list/list";

const improvements = [
  {
    text: "Suggestion 1",
    link: "/suggestion1",
  },
  {
    text: "Suggestion 2 has more text",
    link: "/suggestion2",
  },
  {
    text: "Suggestion 3 has a lot of text it just keeps going it is such a long suggestion because there is so much to say.",
    link: "/suggestion3",
  },
];

export const SecurityScoreCard = ({
  score,
}) => {
  return (
    <DefaultWrapper width={450}>
      <DefaultContainer>
        <Card
          topRadius
          size="l"
          label={"Security score"}
          count={score}
          fullWidth
          isScore
        />
        <SuggestionsList suggestions={improvements} />
      </DefaultContainer>
    </DefaultWrapper>
  );
};
