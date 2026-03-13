import React from "react";
import { Box, Button, Typography } from "@mui/material";
import { normalizeString } from "../../../common/helpers";
import pluralize from "pluralize";

interface DrilldownPaginationProps {
  memberType: string;
  startIndex: number;
  endIndex: number;
  totalCount: number;
  hasPrev: boolean;
  hasNext: boolean;
  onPrev: () => void;
  onNext: () => void;
}

export const DrilldownPagination: React.FC<DrilldownPaginationProps> = ({
  memberType,
  startIndex,
  endIndex,
  totalCount,
  hasPrev,
  hasNext,
  onPrev,
  onNext,
}) => {
  const typeName = pluralize(normalizeString(memberType, true));
  return (
    <Box
      sx={{
        px: 2,
        py: 1,
        borderTop: 1,
        borderColor: "divider",
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        flexShrink: 0,
      }}
    >
      <Button size="small" disabled={!hasPrev} onClick={onPrev}>
        Prev
      </Button>
      <Typography variant="caption">
        Showing {startIndex}-{endIndex} of {totalCount} {typeName}
      </Typography>
      <Button size="small" disabled={!hasNext} onClick={onNext}>
        Next
      </Button>
    </Box>
  );
};
