import React, { useCallback, useMemo, useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TableSortLabel,
  Typography,
  Chip,
  Box,
  CircularProgress,
} from "@mui/material";
import { FixedSizeList } from "react-window";
import InfiniteLoader from "react-window-infinite-loader";
import { normalizeString } from "../../../common/helpers";

type SortDirection = "asc" | "desc";

interface ResourceTableProps {
  data: any[];
  isUserTrait: boolean;
  hasMore: boolean;
  loading: boolean;
  onLoadMore: () => Promise<void>;
  onRowClick: (item: any) => void;
  countsByType?: Record<string, number>;
  totalCount?: number;
}

const ROW_HEIGHT = 44;

export const ResourceTable: React.FC<ResourceTableProps> = ({
  data,
  isUserTrait,
  hasMore,
  loading,
  onLoadMore,
  onRowClick,
  countsByType,
  totalCount,
}) => {
  const [sortField, setSortField] = useState<string>("name");
  const [sortDirection, setSortDirection] = useState<SortDirection>("asc");

  const handleSort = (field: string) => {
    if (sortField === field) {
      setSortDirection((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortField(field);
      setSortDirection("asc");
    }
  };

  const getName = (item: any): string => {
    return item.resource?.display_name || "";
  };

  const getType = (item: any): string => {
    return item.resource_type?.id || "";
  };

  const getEntitlements = (item: any): string[] => {
    const ents = item.entitlements || [];
    return ents.map((e: any) => e.display_name || e.slug || "").filter(Boolean);
  };

  // Show department and job_title profile columns when available
  const PROFILE_COLUMNS = ["department", "job_title"];
  const profileKeys = useMemo(() => {
    const available = new Set<string>();
    for (const item of data) {
      if (item.profile) {
        for (const k of Object.keys(item.profile)) {
          available.add(k);
        }
      }
    }
    return PROFILE_COLUMNS.filter((k) => available.has(k));
  }, [data]);

  const sortedData = useMemo(() => {
    const sorted = [...data];
    sorted.sort((a, b) => {
      let aVal: string, bVal: string;
      if (sortField === "name") {
        aVal = getName(a).toLowerCase();
        bVal = getName(b).toLowerCase();
      } else if (sortField === "type") {
        aVal = getType(a).toLowerCase();
        bVal = getType(b).toLowerCase();
      } else {
        // Profile column sort
        aVal = (a.profile?.[sortField] || "").toLowerCase();
        bVal = (b.profile?.[sortField] || "").toLowerCase();
      }
      const cmp = aVal.localeCompare(bVal);
      return sortDirection === "asc" ? cmp : -cmp;
    });
    return sorted;
  }, [data, sortField, sortDirection]);

  const itemCount = hasMore ? sortedData.length + 1 : sortedData.length;
  const isItemLoaded = (index: number) => !hasMore || index < sortedData.length;

  const loadMoreItems = useCallback(async () => {
    await onLoadMore();
  }, [onLoadMore]);

  const showSummaryChips = !isUserTrait && countsByType && totalCount && totalCount > 100;

  // Compute flex values: fixed columns get base flex, profile columns share remaining space
  const nameFlex = 2;
  const typeFlex = 1;
  const entFlex = 2;
  const profileFlex = profileKeys.length > 0 ? 1 : 0;

  const Row = ({ index, style }: { index: number; style: React.CSSProperties }) => {
    if (!isItemLoaded(index)) {
      return (
        <div style={{ ...style, display: "flex", alignItems: "center", justifyContent: "center" }}>
          <CircularProgress color="success" size={20} />
        </div>
      );
    }

    const item = sortedData[index];
    const entitlements = getEntitlements(item);

    return (
      <TableRow
        component="div"
        hover
        onClick={() => onRowClick(item)}
        style={{ ...style, display: "flex", cursor: "pointer", alignItems: "center" }}
      >
        <TableCell
          component="div"
          sx={{ flex: nameFlex, border: "none", py: 0, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}
        >
          <Typography variant="body2" noWrap>
            {getName(item)}
          </Typography>
        </TableCell>
        <TableCell
          component="div"
          sx={{ flex: typeFlex, border: "none", py: 0, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}
        >
          <Typography variant="body2" color="text.secondary" noWrap>
            {normalizeString(getType(item), true)}
          </Typography>
        </TableCell>
        {profileKeys.map((key) => (
          <TableCell
            key={key}
            component="div"
            sx={{ flex: profileFlex, border: "none", py: 0, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}
          >
            <Typography variant="body2" color="text.secondary" noWrap>
              {item.profile?.[key] || ""}
            </Typography>
          </TableCell>
        ))}
        <TableCell
          component="div"
          sx={{ flex: entFlex, border: "none", py: 0, overflow: "hidden" }}
        >
          <Box sx={{ display: "flex", gap: 0.5, flexWrap: "nowrap", overflow: "hidden" }}>
            {entitlements.slice(0, 3).map((e, i) => (
              <Chip key={i} label={e} size="small" sx={{ fontSize: "11px", height: 22 }} />
            ))}
            {entitlements.length > 3 && (
              <Chip label={`+${entitlements.length - 3}`} size="small" variant="outlined" sx={{ fontSize: "11px", height: 22 }} />
            )}
          </Box>
        </TableCell>
      </TableRow>
    );
  };

  if (loading && data.length === 0) {
    return (
      <Box sx={{ display: "flex", justifyContent: "center", p: 4 }}>
        <CircularProgress color="success" />
      </Box>
    );
  }

  if (data.length === 0) {
    return (
      <Box sx={{ p: 4, textAlign: "center" }}>
        <Typography color="text.secondary">No data available</Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ display: "flex", flexDirection: "column", flex: 1, minHeight: 0 }}>
      {showSummaryChips && (
        <Box sx={{ display: "flex", gap: 1, px: 2, py: 1, flexWrap: "wrap" }}>
          <Chip
            label={`${totalCount} total members`}
            size="small"
            color="primary"
            variant="outlined"
          />
          {Object.entries(countsByType).map(([typeId, count]) => (
            <Chip
              key={typeId}
              label={`${count} ${normalizeString(typeId, true)}`}
              size="small"
              variant="outlined"
            />
          ))}
        </Box>
      )}
      <TableContainer component="div" sx={{ flex: 1, minHeight: 0 }}>
        <Table component="div" size="small" sx={{ tableLayout: "fixed" }}>
          <TableHead component="div">
            <TableRow component="div" sx={{ display: "flex" }}>
              <TableCell component="div" sx={{ flex: nameFlex }}>
                <TableSortLabel
                  active={sortField === "name"}
                  direction={sortField === "name" ? sortDirection : "asc"}
                  onClick={() => handleSort("name")}
                >
                  {isUserTrait ? "Name" : "Principal Name"}
                </TableSortLabel>
              </TableCell>
              <TableCell component="div" sx={{ flex: typeFlex }}>
                <TableSortLabel
                  active={sortField === "type"}
                  direction={sortField === "type" ? sortDirection : "asc"}
                  onClick={() => handleSort("type")}
                >
                  {isUserTrait ? "Type" : "Principal Type"}
                </TableSortLabel>
              </TableCell>
              {profileKeys.map((key) => (
                <TableCell key={key} component="div" sx={{ flex: profileFlex }}>
                  <TableSortLabel
                    active={sortField === key}
                    direction={sortField === key ? sortDirection : "asc"}
                    onClick={() => handleSort(key)}
                  >
                    {normalizeString(key, true)}
                  </TableSortLabel>
                </TableCell>
              ))}
              <TableCell component="div" sx={{ flex: entFlex }}>
                Entitlements
              </TableCell>
            </TableRow>
          </TableHead>
        </Table>
        <InfiniteLoader
          isItemLoaded={isItemLoaded}
          itemCount={itemCount}
          loadMoreItems={loadMoreItems}
          threshold={10}
        >
          {({ onItemsRendered, ref }) => (
            <FixedSizeList
              height={Math.max(200, window.innerHeight - 300)}
              width="100%"
              itemCount={itemCount}
              itemSize={ROW_HEIGHT}
              onItemsRendered={onItemsRendered}
              ref={ref}
            >
              {Row}
            </FixedSizeList>
          )}
        </InfiniteLoader>
      </TableContainer>
    </Box>
  );
};
