import React, { useEffect, useMemo, useState } from "react";
import { Resources } from "./resourcesGraph/resourcesGraph";
import { useResources } from "../../../context/resources";
import { fetchResourcesWithUserCount } from "../../../components/explorer/api";
import {
  Box,
  Chip,
  CircularProgress,
  InputAdornment,
  Link as MuiLink,
  Paper,
  Skeleton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TableSortLabel,
  TextField,
  Typography,
} from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import ClearIcon from "@mui/icons-material/Clear";
import ArrowForwardIcon from "@mui/icons-material/ArrowForward";
import PeopleIcon from "@mui/icons-material/PeopleOutline";
import { Link } from "react-router-dom";
import { normalizeString } from "../../../common/helpers";
import { isObjectEmpty } from "../../../common/helpers";
import pluralize from "pluralize";
import { colors } from "../../../style/colors";

type SortField = "name" | "type" | "members";
type SortDir = "asc" | "desc";

export const Inventory = () => {
  const { identities, groupTraitTypes, roleTraitTypes, loading: resourcesLoading } = useResources();
  const [roles, setRoles] = useState<Record<string, any[]>>({});
  const [groups, setGroups] = useState<Record<string, any[]>>({});
  const [countsLoading, setCountsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [sortField, setSortField] = useState<SortField>("members");
  const [sortDir, setSortDir] = useState<SortDir>("desc");

  useEffect(() => {
    if (groupTraitTypes.length === 0 && roleTraitTypes.length === 0) return;

    const fetchData = async () => {
      const roleData: Record<string, any[]> = {};
      for (const type of roleTraitTypes) {
        try {
          const response = await fetchResourcesWithUserCount(type);
          roleData[type] = response?.resources || [];
        } catch {
          roleData[type] = [];
        }
      }

      const groupData: Record<string, any[]> = {};
      for (const type of groupTraitTypes) {
        try {
          const response = await fetchResourcesWithUserCount(type);
          groupData[type] = response?.resources || [];
        } catch {
          groupData[type] = [];
        }
      }

      setRoles(roleData);
      setGroups(groupData);
      setCountsLoading(false);
    };
    fetchData();
  }, [groupTraitTypes, roleTraitTypes]);

  const allResources = useMemo(() => {
    const items: { name: string; type: string; typeLabel: string; members: number; resourceTypeId: string; resourceId: string }[] = [];
    const addItems = (data: Record<string, any[]>, traitLabel: string) => {
      for (const [typeId, resources] of Object.entries(data)) {
        for (const r of resources) {
          items.push({
            name: r.resource?.display_name || "",
            type: typeId,
            typeLabel: `${normalizeString(typeId, true)} (${traitLabel})`,
            members: r.userCount || 0,
            resourceTypeId: r.resource_type?.id || typeId,
            resourceId: r.resource?.id?.resource || "",
          });
        }
      }
    };
    addItems(roles, "Role");
    addItems(groups, "Group");
    return items;
  }, [roles, groups]);

  const filteredResources = useMemo(() => {
    let items = allResources;
    if (search.trim()) {
      const q = search.toLowerCase();
      items = items.filter(
        (r) => r.name.toLowerCase().includes(q) || r.type.toLowerCase().includes(q) || r.typeLabel.toLowerCase().includes(q)
      );
    }
    items.sort((a, b) => {
      let cmp: number;
      if (sortField === "name") {
        cmp = a.name.localeCompare(b.name);
      } else if (sortField === "type") {
        cmp = a.typeLabel.localeCompare(b.typeLabel);
      } else {
        cmp = a.members - b.members;
      }
      return sortDir === "asc" ? cmp : -cmp;
    });
    return items;
  }, [allResources, search, sortField, sortDir]);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortField(field);
      setSortDir(field === "members" ? "desc" : "asc");
    }
  };

  if (resourcesLoading) {
    return (
      <Box sx={{ display: "flex", justifyContent: "center", alignItems: "center", width: "100%", p: 4 }}>
        <CircularProgress color="success" />
      </Box>
    );
  }

  return (
    <Box sx={{ display: "flex", flexDirection: "column", width: "100%", gap: 2, p: 2 }}>
      {/* Summary stats bar */}
      <Box sx={{ display: "flex", gap: 1.5, alignItems: "center", flexWrap: "wrap" }}>
        <StatChip
          label="Identities"
          count={identities.count}
        />
        {identities.identityTypes?.map((type) => (
          <StatChip
            key={type}
            label={pluralize(normalizeString(type, true))}
            count={identities.resourcesByType?.[type]?.length || 0}
            to={`/${type}`}
          />
        ))}
        {roleTraitTypes.map((type) => (
          <StatChip
            key={type}
            label={pluralize(normalizeString(type, true))}
            count={roles[type]?.length}
            to={`/${type}`}
            loading={countsLoading}
          />
        ))}
        {groupTraitTypes.map((type) => (
          <StatChip
            key={type}
            label={pluralize(normalizeString(type, true))}
            count={groups[type]?.length}
            to={`/${type}`}
            loading={countsLoading}
          />
        ))}
      </Box>

      {/* Main content: table + chart side by side */}
      <Box sx={{ display: "flex", gap: 2, flex: 1, minHeight: 0, alignItems: "flex-start" }}>
        {/* Resources table */}
        <Paper variant="outlined" sx={{ flex: 1, minWidth: 0, overflow: "hidden" }}>
          <Box sx={{ display: "flex", alignItems: "center", justifyContent: "space-between", px: 2, py: 1.5 }}>
            <Typography variant="subtitle1" fontWeight={600}>
              Roles & Groups
            </Typography>
            <TextField
              size="small"
              placeholder="Filter..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              sx={{ width: 220 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon sx={{ fontSize: 18 }} />
                  </InputAdornment>
                ),
                endAdornment: search ? (
                  <InputAdornment position="end">
                    <ClearIcon
                      sx={{ fontSize: 18, cursor: "pointer" }}
                      onClick={() => setSearch("")}
                    />
                  </InputAdornment>
                ) : null,
              }}
            />
          </Box>
          {countsLoading ? (
            <Box sx={{ p: 2 }}>
              {[1, 2, 3, 4, 5].map((i) => (
                <Skeleton key={i} variant="rectangular" height={36} sx={{ mb: 0.5, borderRadius: 1 }} />
              ))}
            </Box>
          ) : (
            <TableContainer sx={{ maxHeight: "calc(100vh - 260px)" }}>
              <Table size="small" stickyHeader>
                <TableHead>
                  <TableRow>
                    <TableCell>
                      <TableSortLabel
                        active={sortField === "name"}
                        direction={sortField === "name" ? sortDir : "asc"}
                        onClick={() => handleSort("name")}
                      >
                        Name
                      </TableSortLabel>
                    </TableCell>
                    <TableCell>
                      <TableSortLabel
                        active={sortField === "type"}
                        direction={sortField === "type" ? sortDir : "asc"}
                        onClick={() => handleSort("type")}
                      >
                        Type
                      </TableSortLabel>
                    </TableCell>
                    <TableCell align="right" sx={{ width: 120 }}>
                      <TableSortLabel
                        active={sortField === "members"}
                        direction={sortField === "members" ? sortDir : "desc"}
                        onClick={() => handleSort("members")}
                      >
                        Members
                      </TableSortLabel>
                    </TableCell>
                    <TableCell sx={{ width: 48 }} />
                  </TableRow>
                </TableHead>
                <TableBody>
                  {filteredResources.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={4} align="center" sx={{ py: 3 }}>
                        <Typography color="text.secondary" variant="body2">
                          {search ? "No matching resources" : "No roles or groups found"}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  ) : (
                    filteredResources.map((r) => (
                      <TableRow
                        key={`${r.resourceTypeId}-${r.resourceId}`}
                        hover
                        component={Link}
                        to={`/${r.resourceTypeId}/${r.resourceId}`}
                        state={{ from: "/dashboard" }}
                        sx={{ textDecoration: "none", cursor: "pointer" }}
                      >
                        <TableCell>
                          <Typography variant="body2" noWrap>
                            {r.name}
                          </Typography>
                        </TableCell>
                        <TableCell>
                          <Typography variant="body2" color="text.secondary" noWrap>
                            {r.typeLabel}
                          </Typography>
                        </TableCell>
                        <TableCell align="right">
                          <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 0.5 }}>
                            <PeopleIcon sx={{ fontSize: 16, color: "text.secondary" }} />
                            <Typography variant="body2" fontWeight={500}>
                              {r.members.toLocaleString()}
                            </Typography>
                          </Box>
                        </TableCell>
                        <TableCell>
                          <ArrowForwardIcon sx={{ fontSize: 16, color: "text.secondary" }} />
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Paper>

        {/* Resource distribution chart */}
        <Box sx={{ width: 340, flexShrink: 0 }}>
          <Resources />
        </Box>
      </Box>
    </Box>
  );
};

const StatChip = ({
  label,
  count,
  to,
  loading,
}: {
  label: string;
  count?: number;
  to?: string;
  loading?: boolean;
}) => {
  const chip = (
    <Chip
      label={
        <Box sx={{ display: "flex", alignItems: "center", gap: 0.75 }}>
          <Typography variant="body2" fontWeight={600}>
            {loading ? "..." : (count ?? 0).toLocaleString()}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {label}
          </Typography>
        </Box>
      }
      variant="outlined"
      size="medium"
      clickable={!!to}
      sx={{
        borderRadius: "8px",
        height: 36,
        "& .MuiChip-label": { px: 1.5 },
      }}
    />
  );

  if (to) {
    return (
      <MuiLink component={Link} to={to} underline="none">
        {chip}
      </MuiLink>
    );
  }
  return chip;
};
