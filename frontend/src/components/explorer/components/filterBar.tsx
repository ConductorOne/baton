import React, { useState } from "react";
import {
  Box,
  Button,
  IconButton,
  InputAdornment,
  MenuItem,
  Popover,
  Select,
  TextField,
  Typography,
} from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import ClearIcon from "@mui/icons-material/Clear";
import CodeIcon from "@mui/icons-material/Code";
import HelpOutlineIcon from "@mui/icons-material/HelpOutline";
import { normalizeString } from "../../../common/helpers";

export interface FilterState {
  textSearch: string;
  typeFilter: string;
  entitlementSearch: string;
  celExpression: string;
  celActive: boolean;
}

interface FilterBarProps {
  filterState: FilterState;
  onFilterChange: (state: FilterState) => void;
  countsByType?: Record<string, number>;
  isUserTrait: boolean;
  celError?: string;
  celLoading?: boolean;
  onCelApply: (expression: string) => void;
  onCelClear: () => void;
}

export const createEmptyFilterState = (): FilterState => ({
  textSearch: "",
  typeFilter: "",
  entitlementSearch: "",
  celExpression: "",
  celActive: false,
});

export const FilterBar: React.FC<FilterBarProps> = ({
  filterState,
  onFilterChange,
  countsByType,
  isUserTrait,
  celError,
  celLoading,
  onCelApply,
  onCelClear,
}) => {
  const [showCel, setShowCel] = useState(false);
  const [helpAnchor, setHelpAnchor] = useState<HTMLElement | null>(null);

  const handleTextChange = (value: string) => {
    onFilterChange({ ...filterState, textSearch: value });
  };

  const handleTypeChange = (value: string) => {
    onFilterChange({ ...filterState, typeFilter: value });
  };

  const handleEntitlementChange = (value: string) => {
    onFilterChange({ ...filterState, entitlementSearch: value });
  };

  const handleCelExpressionChange = (value: string) => {
    onFilterChange({ ...filterState, celExpression: value });
  };

  const handleCelKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && filterState.celExpression.trim()) {
      onCelApply(filterState.celExpression);
    }
  };

  const handleCelClear = () => {
    onFilterChange({ ...filterState, celExpression: "", celActive: false });
    onCelClear();
  };

  const typeOptions = countsByType ? Object.keys(countsByType) : [];

  const celScope = isUserTrait ? "resources" : "grants";
  const celVariables = isUserTrait
    ? ["resource.display_name", "resource.id", "resource.type", 'resource.profile (map, e.g. resource.profile["email"])']
    : ["principal.display_name", "principal.type", "principal.id", "entitlement.display_name", "entitlement.slug", 'principal.profile (map, e.g. principal.profile["email"])'];
  const celExample = isUserTrait
    ? 'resource.profile["email"].contains("@example.com")'
    : 'principal.profile["department"] == "Engineering"';

  return (
    <Box sx={{ px: 2, py: 1, display: "flex", flexDirection: "column", gap: 1 }}>
      <Box sx={{ display: "flex", gap: 1, alignItems: "center", flexWrap: "wrap" }}>
        <TextField
          size="small"
          placeholder="Search name, type, entitlement, profile..."
          value={filterState.textSearch}
          onChange={(e) => handleTextChange(e.target.value)}
          sx={{ minWidth: 200, flex: 1, maxWidth: 400 }}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon sx={{ fontSize: 18 }} />
              </InputAdornment>
            ),
            endAdornment: filterState.textSearch ? (
              <InputAdornment position="end">
                <ClearIcon
                  sx={{ fontSize: 18, cursor: "pointer" }}
                  onClick={() => handleTextChange("")}
                />
              </InputAdornment>
            ) : null,
          }}
        />
        <TextField
          size="small"
          placeholder="Filter by entitlement..."
          value={filterState.entitlementSearch}
          onChange={(e) => handleEntitlementChange(e.target.value)}
          sx={{ minWidth: 180, maxWidth: 300 }}
          InputProps={{
            endAdornment: filterState.entitlementSearch ? (
              <InputAdornment position="end">
                <ClearIcon
                  sx={{ fontSize: 18, cursor: "pointer" }}
                  onClick={() => handleEntitlementChange("")}
                />
              </InputAdornment>
            ) : null,
          }}
        />
        {typeOptions.length > 0 && (
          <Select
            size="small"
            value={filterState.typeFilter}
            onChange={(e) => handleTypeChange(e.target.value)}
            displayEmpty
            sx={{ minWidth: 160 }}
          >
            <MenuItem value="">All Types</MenuItem>
            {typeOptions.map((typeId) => (
              <MenuItem key={typeId} value={typeId}>
                {normalizeString(typeId, true)} ({countsByType[typeId]})
              </MenuItem>
            ))}
          </Select>
        )}
        <Button
          size="small"
          variant={showCel ? "contained" : "outlined"}
          startIcon={<CodeIcon />}
          onClick={() => setShowCel((v) => !v)}
          sx={{ textTransform: "none" }}
        >
          CEL Filter
        </Button>
      </Box>

      {showCel && (
        <Box sx={{ display: "flex", gap: 1, alignItems: "flex-start" }}>
          <TextField
            size="small"
            fullWidth
            placeholder={celExample}
            value={filterState.celExpression}
            onChange={(e) => handleCelExpressionChange(e.target.value)}
            onKeyDown={handleCelKeyDown}
            error={!!celError}
            helperText={celError}
            disabled={celLoading}
            InputProps={{
              sx: { fontFamily: "monospace", fontSize: 13 },
            }}
          />
          <Button
            size="small"
            variant="contained"
            onClick={() => onCelApply(filterState.celExpression)}
            disabled={!filterState.celExpression.trim() || celLoading}
            sx={{ textTransform: "none", whiteSpace: "nowrap" }}
          >
            {celLoading ? "..." : "Apply"}
          </Button>
          {filterState.celActive && (
            <Button
              size="small"
              variant="outlined"
              onClick={handleCelClear}
              sx={{ textTransform: "none" }}
            >
              Clear
            </Button>
          )}
          <IconButton size="small" onClick={(e) => setHelpAnchor(e.currentTarget)}>
            <HelpOutlineIcon fontSize="small" />
          </IconButton>
          <Popover
            open={!!helpAnchor}
            anchorEl={helpAnchor}
            onClose={() => setHelpAnchor(null)}
            anchorOrigin={{ vertical: "bottom", horizontal: "left" }}
          >
            <Box sx={{ p: 2, maxWidth: 360 }}>
              <Typography variant="subtitle2" gutterBottom>
                CEL Filter ({celScope} scope)
              </Typography>
              <Typography variant="body2" gutterBottom>
                Available variables:
              </Typography>
              <Box component="ul" sx={{ m: 0, pl: 2 }}>
                {celVariables.map((v) => (
                  <li key={v}>
                    <Typography variant="body2" sx={{ fontFamily: "monospace", fontSize: 12 }}>
                      {v}
                    </Typography>
                  </li>
                ))}
              </Box>
              <Typography variant="body2" sx={{ mt: 1, fontFamily: "monospace", fontSize: 12 }}>
                Example: {celExample}
              </Typography>
            </Box>
          </Popover>
        </Box>
      )}
    </Box>
  );
};
