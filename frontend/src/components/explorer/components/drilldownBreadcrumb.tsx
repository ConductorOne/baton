import React from "react";
import { Breadcrumbs, Link, Typography, Box } from "@mui/material";
import { normalizeString } from "../../../common/helpers";
import pluralize from "pluralize";

interface DrilldownBreadcrumbProps {
  resourceName: string;
  memberType?: string;
  memberTypeCount?: number;
  onBackToSummary: () => void;
}

export const DrilldownBreadcrumb: React.FC<DrilldownBreadcrumbProps> = ({
  resourceName,
  memberType,
  memberTypeCount,
  onBackToSummary,
}) => {
  return (
    <Box sx={{ px: 2, py: 1, borderBottom: 1, borderColor: "divider", flexShrink: 0 }}>
      <Breadcrumbs>
        {memberType ? (
          <Link
            component="button"
            underline="hover"
            color="inherit"
            onClick={onBackToSummary}
            sx={{ cursor: "pointer" }}
          >
            {resourceName}
          </Link>
        ) : (
          <Typography color="text.primary">{resourceName}</Typography>
        )}
        {memberType && (
          <Typography color="text.primary">
            {pluralize(normalizeString(memberType, true))} ({memberTypeCount})
          </Typography>
        )}
      </Breadcrumbs>
    </Box>
  );
};
