import * as React from "react";
import { useEffect, useState } from "react";
import { PieTypeChart } from "../../components/charts/pieChart";
import { colors } from "../../../../style/colors";
import {
  Box,
  CircularProgress,
  List,
  ListItem,
  ListItemText,
  Paper,
  Typography,
  useTheme,
} from "@mui/material";
import { Link } from "react-router-dom";
import { useResources } from "../../../../context/resources";
import pluralize from "pluralize";
import { normalizeString } from "../../../../common/helpers";
import { fetchResourceTypes } from "../../../../components/explorer/api";
import { fetchResourcesByType } from "../../../../components/explorer/api";

const graphColors = [
  colors.batonGreen600,
  colors.batonGreen800,
  colors.pink300,
  colors.blue500,
  colors.purple400,
  colors.blue200,
  colors.cyan400,
  colors.indigo500,
];

export const Resources = () => {
  const theme = useTheme();
  const lightMode = theme.palette.mode === "light";
  const [graphData, setGraphData] = useState<any[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      try {
        const rtData = await fetchResourceTypes();
        const resourceTypes = rtData?.resource_types || [];

        let total = 0;
        const data: any[] = [];

        for (const rtOutput of resourceTypes) {
          const rt = rtOutput.resource_type;
          if (!rt?.id) continue;

          const resp = await fetchResourcesByType(rt.id);
          const count = resp.total_count || resp.data?.resources?.length || 0;

          if (count > 0) {
            data.push({
              name: pluralize(normalizeString(rt.id, false)),
              value: count,
              type: rt.id,
            });
          }
          total += count;
        }

        setGraphData(data);
        setTotalCount(total);
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, []);

  if (loading) {
    return (
      <Paper variant="outlined" sx={{ p: 2, display: "flex", justifyContent: "center" }}>
        <CircularProgress color="success" size={24} />
      </Paper>
    );
  }

  return (
    <Paper variant="outlined" sx={{ overflow: "hidden" }}>
      <Box sx={{ px: 2, pt: 1.5, pb: 0.5 }}>
        <Typography variant="subtitle1" fontWeight={600}>
          Resources
        </Typography>
      </Box>
      <Box sx={{ display: "flex", justifyContent: "center", py: 1 }}>
        <PieTypeChart
          data={graphData}
          colors={graphColors}
          width={300}
          height={180}
          textPositionX={150}
          textPositionY={100}
          textFillColor={
            lightMode ? colors.batonGreen900 : colors.batonGreen500
          }
          textSize={36}
          text={totalCount}
        />
      </Box>
      <List dense disablePadding>
        {graphData.map((d, i) => (
          <ListItem
            key={d.type}
            component={Link}
            to={`/${d.type}`}
            sx={{
              textDecoration: "none",
              color: "text.primary",
              px: 2,
              py: 0.5,
              "&:hover": {
                backgroundColor: lightMode ? colors.gray50 : "rgba(255,255,255,0.04)",
              },
            }}
          >
            <Box
              sx={{
                width: 10,
                height: 10,
                borderRadius: "50%",
                backgroundColor: graphColors[i % graphColors.length],
                mr: 1.5,
                flexShrink: 0,
              }}
            />
            <ListItemText
              primary={d.name}
              primaryTypographyProps={{ variant: "body2", noWrap: true }}
            />
            <Typography variant="body2" fontWeight={600} sx={{ ml: 1 }}>
              {d.value.toLocaleString()}
            </Typography>
          </ListItem>
        ))}
      </List>
    </Paper>
  );
};
