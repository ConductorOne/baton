import React, { useEffect, useState, useCallback, useMemo, useRef } from "react";
import ReactFlow, {
  Controls,
  useEdgesState,
  useNodesInitialized,
  useNodesState,
  useReactFlow,
  getOutgoers,
  getIncomers,
  getConnectedEdges,
} from "reactflow";
import "reactflow/dist/style.css";
import { CustomEdge } from "./components/customEdge";
import {
  AggregateNode,
  ChildNode,
  EntitlementNode,
  ExpandableGrantNode,
  ParentNode,
} from "./components/customNode";
import { ResourcesSidebar } from "./components/resourcesSidebar";
import { ResourceTable } from "./components/resourceTable";
import { FilterBar, FilterState, createEmptyFilterState } from "./components/filterBar";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import { ExplorerLayout, TreeWrapper } from "./styles/styles";
import { ResourceDetailsModal } from "../resourceDetails";
import { extractPrincipalId, isObjectEmpty } from "../../common/helpers";
import {
  fetchAccessForUser,
  fetchGrantsForResource,
  fetchResourceDetails,
  searchWithCEL,
} from "./api";
import { populateNodes } from "./nodesAndEdges";
import { Box, ToggleButton, ToggleButtonGroup, Tooltip, useTheme } from "@mui/material";
import TableChartIcon from "@mui/icons-material/TableChart";
import AccountTreeIcon from "@mui/icons-material/AccountTree";
import { colors } from "../../style/colors";

type ResourceDetailsState = {
  resourceOpened?: boolean;
  entitlementOpened?: boolean;
  resource?: any;
};

type ViewMode = "table" | "graph";

type TableDataContext = {
  data: any[];
  isUserTrait: boolean;
  hasMore: boolean;
  loading: boolean;
  nextPageToken?: string;
  resourceType?: string;
  resourceId?: string;
  totalCount?: number;
  countsByType?: Record<string, number>;
  parentResource?: any;
};

const edgeTypes = { customEdge: CustomEdge };
const nodeTypes = {
  parent: ParentNode,
  child: ChildNode,
  expandable: ExpandableGrantNode,
  aggregate: AggregateNode,
  entitlement: EntitlementNode,
};

const Explorer = ({ resourceList, closeResourceList }) => {
  const navigate = useNavigate();
  const { state } = useLocation();
  const fromDashboard = state?.from === "/dashboard";
  const theme = useTheme();
  const reactFlowInstance = useReactFlow();
  const nodesInitialized = useNodesInitialized();
  const { id: resourceId } = useParams();
  const { type: resourceType } = useParams();
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [viewMode, setViewMode] = useState<ViewMode>("table");
  const [tableData, setTableData] = useState<TableDataContext | null>(null);
  const [filterState, setFilterState] = useState<FilterState>(createEmptyFilterState());
  const [celError, setCelError] = useState<string>("");
  const [celLoading, setCelLoading] = useState(false);
  const [celResults, setCelResults] = useState<any[] | null>(null);
  const [resourceDetailsOpen, setResourceDetailsOpen] =
    useState<ResourceDetailsState>({
      resourceOpened: false,
      entitlementOpened: false,
    });
  const [detailsPosition, setDetailsPosition] = useState<{ top: number; left: number } | null>(null);
  const fetchingRef = useRef(false);
  const tableNextPageTokenRef = useRef<string | undefined>(undefined);
  const tableLoadingMoreRef = useRef(false);

  useEffect(() => {
    if (resourceId && !fetchingRef.current) {
      fetchingRef.current = true;
      const fetchData = async () => {
        const resource = await fetchResourceDetails(resourceType, resourceId);
        if (resource) {
          await openTreeView(resource, true);
        }
        fetchingRef.current = false;
      };
      fetchData();
    }
  }, [resourceId, resourceType]);

  useEffect(() => {
    if (nodesInitialized && viewMode === "graph") {
      reactFlowInstance.fitView();
    }
  }, [nodesInitialized, viewMode]);

  const openEntitlementsDetails = (entitlement, position?: { top: number; left: number }) => {
    if (position) setDetailsPosition(position);
    setResourceDetailsOpen({
      entitlementOpened: true,
      resource: entitlement,
    });
  };

  const openResourceDetails = async (
    resourceType: string,
    principalId: string,
    currentNode,
    clickPosition?: { top: number; left: number }
  ) => {
    // For entitlement nodes, show entitlement details
    if (currentNode.data?.isEntitlement) {
      openEntitlementsDetails(currentNode.data.entitlement, clickPosition);
      return;
    }

    if (clickPosition) setDetailsPosition(clickPosition);

    const resourceDetails = await fetchResourceDetails(
      resourceType,
      principalId
    );
    setResourceDetailsOpen({
      resourceOpened: true,
      resource: resourceDetails,
    });

    const connectedEdges = getConnectedEdges([currentNode], edges);

    if (connectedEdges.length > 0) {
      connectedEdges.forEach((edge) => {
        edge.data.style = { opacity: 0.3 };
      });
    }

    edges?.forEach((edge) => {
      edge.data.style = { opacity: 0.3 };
      if (connectedEdges.includes(edge)) {
        edge.data.style = { opacity: 1 };
      }
    });

    const outgoers = getOutgoers(currentNode, nodes, edges);
    const incomers = getIncomers(currentNode, nodes, edges);

    const connectedNodes = [];
    if (outgoers.length > 0) {
      connectedNodes.push(...outgoers);
    }

    if (incomers.length > 0) {
      connectedNodes.push(...incomers);
    }

    setNodes((nds) =>
      nds.map((node) => {
        node.style = { opacity: "0.3" };

        if (connectedNodes.includes(node)) {
          node.style = {
            border: `1.2px solid ${
              theme.palette.mode === "light"
                ? colors.batonGreen600
                : colors.batonGreen500
            }`,
            borderRadius: "12px",
            opacity: 1,
          };
        }
        if (node.selected) {
          node.style = { opacity: 1 };
        }
        return node;
      })
    );
  };

  const closeResourceDetails = () => {
    edges.forEach((e) => {
      e.data.style = {};
    });

    setNodes((nds) =>
      nds.map((node) => {
        node.style = { opacity: 1 };
        node.selected = false;
        return node;
      })
    );

    setResourceDetailsOpen({
      resourceOpened: false,
      entitlementOpened: false,
    });
    setDetailsPosition(null);
  };

  const openTreeView = async (resource, skipNavigate = false) => {
    // Prevent the useEffect from triggering a duplicate fetch when navigate() changes URL params
    fetchingRef.current = true;

    if (!skipNavigate && !fromDashboard) {
      navigate(
        `/${resource.resource_type.id}/${resource.resource.id.resource}`
      );
    }

    // Reset filter/CEL state on resource change
    setFilterState(createEmptyFilterState());
    setCelResults(null);
    setCelError("");

    const isUserTrait = resource.resource_type.traits && resource.resource_type.traits[0] === 1;

    if (isUserTrait) {
      const grantsAccess = await fetchAccessForUser(
        resource.resource_type.id,
        resource.resource.id.resource
      );
      if (!grantsAccess.access || !isObjectEmpty(grantsAccess.access)) {
        populateNodes(
          setNodes,
          setEdges,
          grantsAccess,
          resource,
          openEntitlementsDetails
        );
      }

      // Populate table data for user-trait resources
      const accessItems = grantsAccess.access || [];
      setTableData({
        data: accessItems,
        isUserTrait: true,
        hasMore: false,
        loading: false,
        resourceType: resource.resource_type.id,
        resourceId: resource.resource.id.resource,
        parentResource: resource,
      });
      tableNextPageTokenRef.current = undefined;
    } else {
      const grantsResponse = await fetchGrantsForResource(
        resource.resource_type.id,
        resource.resource.id.resource,
        100,
      );
      const grantsAccess = grantsResponse.data;
      const totalCount = grantsResponse.total_count || 0;
      const countsByType = grantsResponse.counts_by_type || {};

      if (!grantsAccess?.access || isObjectEmpty(grantsAccess.access)) {
        setTableData(null);
        closeResourceDetails();
        fetchingRef.current = false;
        return;
      }

      const accessItems = grantsAccess.access || [];
      tableNextPageTokenRef.current = grantsResponse.next_page_token;

      setTableData({
        data: accessItems,
        isUserTrait: false,
        hasMore: !!grantsResponse.next_page_token,
        loading: false,
        nextPageToken: grantsResponse.next_page_token,
        resourceType: resource.resource_type.id,
        resourceId: resource.resource.id.resource,
        totalCount,
        countsByType,
        parentResource: resource,
      });

      populateNodes(
        setNodes,
        setEdges,
        grantsAccess,
        resource,
        openEntitlementsDetails
      );
    }
    closeResourceDetails();
    fetchingRef.current = false;
  };

  const handleEdgeClick = (e, entitlement) => {
    e.stopPropagation();
    openEntitlementsDetails(entitlement);
  };

  // Table: load more pages for grants
  const handleTableLoadMore = useCallback(async () => {
    if (tableLoadingMoreRef.current || !tableNextPageTokenRef.current || !tableData) return;
    tableLoadingMoreRef.current = true;

    try {
      const resp = await fetchGrantsForResource(
        tableData.resourceType!,
        tableData.resourceId!,
        100,
        tableNextPageTokenRef.current,
      );
      const newAccess = resp.data?.access || [];
      tableNextPageTokenRef.current = resp.next_page_token;

      setTableData((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          data: [...prev.data, ...newAccess],
          hasMore: !!resp.next_page_token,
          nextPageToken: resp.next_page_token,
        };
      });
    } finally {
      tableLoadingMoreRef.current = false;
    }
  }, [tableData?.resourceType, tableData?.resourceId]);

  // Table: row click opens resource details modal
  const handleTableRowClick = useCallback(async (item: any) => {
    const resType = item.resource_type?.id || item.resource?.id?.resource_type;
    const resId = item.resource?.id?.resource;
    if (!resType || !resId) return;

    const resourceDetails = await fetchResourceDetails(resType, resId);
    setResourceDetailsOpen({
      resourceOpened: true,
      resource: resourceDetails,
    });
  }, []);

  // CEL search
  const handleCelApply = useCallback(async (expression: string) => {
    if (!tableData || !expression.trim()) return;
    setCelLoading(true);
    setCelError("");

    try {
      const scope = tableData.isUserTrait ? "resources" : "grants";
      const resp = await searchWithCEL({
        cel: expression,
        scope,
        resourceType: tableData.isUserTrait ? undefined : tableData.resourceType,
        resourceId: tableData.isUserTrait ? undefined : tableData.resourceId,
        resourceTypeId: tableData.isUserTrait ? tableData.resourceType : undefined,
        pageSize: 100,
      });
      const results = tableData.isUserTrait
        ? resp.data?.resources || resp.data?.access || resp.data || []
        : resp.data?.access || resp.data || [];
      setCelResults(Array.isArray(results) ? results : []);
      setFilterState((prev) => ({ ...prev, celActive: true }));
    } catch (err: any) {
      const msg = err?.message || "CEL search failed";
      // Try to extract server error message
      try {
        const parsed = JSON.parse(msg.split(": ").slice(1).join(": "));
        setCelError(parsed.error || msg);
      } catch {
        setCelError(msg);
      }
    } finally {
      setCelLoading(false);
    }
  }, [tableData]);

  const handleCelClear = useCallback(() => {
    setCelResults(null);
    setCelError("");
  }, []);

  // Apply client-side filters to table data
  const filteredTableData = useMemo(() => {
    const source = celResults !== null ? celResults : (tableData?.data || []);
    let filtered = source;

    if (filterState.textSearch.trim()) {
      const q = filterState.textSearch.toLowerCase();
      filtered = filtered.filter((item: any) => {
        const name = (item.resource?.display_name || "").toLowerCase();
        const type = (item.resource_type?.id || "").toLowerCase();
        const entitlements = (item.entitlements || [])
          .map((e: any) => (e.display_name || e.slug || "").toLowerCase())
          .join(" ");
        // Search profile attributes (extracted server-side)
        const profileValues = Object.values(item.profile || {})
          .map((v: any) => String(v).toLowerCase())
          .join(" ");
        return name.includes(q) || type.includes(q) || entitlements.includes(q) || profileValues.includes(q);
      });
    }

    if (filterState.typeFilter) {
      filtered = filtered.filter((item: any) => {
        return item.resource_type?.id === filterState.typeFilter;
      });
    }

    if (filterState.entitlementSearch.trim()) {
      const eq = filterState.entitlementSearch.toLowerCase();
      filtered = filtered.filter((item: any) => {
        const ents = item.entitlements || [];
        return ents.some((e: any) =>
          (e.display_name || "").toLowerCase().includes(eq) ||
          (e.slug || "").toLowerCase().includes(eq)
        );
      });
    }

    return filtered;
  }, [tableData?.data, celResults, filterState.textSearch, filterState.typeFilter, filterState.entitlementSearch]);

  const hasGraphData = nodes.length > 0 && edges.length > 0;
  const hasData = tableData !== null || hasGraphData;

  // Build graph from currently filtered table data
  const buildGraphFromFilteredData = useCallback(() => {
    if (!tableData || filteredTableData.length === 0) return;

    const parentRes = tableData.parentResource;
    if (!parentRes) return;

    if (tableData.isUserTrait) {
      const principal = parentRes.resource || {};
      const fakeAccess = {
        principal: {
          id: principal.id,
          display_name: principal.display_name,
        },
        access: filteredTableData,
      };
      populateNodes(setNodes, setEdges, fakeAccess, parentRes, openEntitlementsDetails);
    } else {
      const fakeAccess = {
        resource: parentRes.resource,
        resource_type: parentRes.resource_type,
        access: filteredTableData,
      };
      populateNodes(setNodes, setEdges, fakeAccess, parentRes, openEntitlementsDetails);
    }
  }, [filteredTableData, tableData, openEntitlementsDetails]);

  return (
    <>
      {resourceList.opened && (
        <ResourcesSidebar
          closeResourceList={closeResourceList}
          resourceType={resourceList.resource}
          openTreeView={openTreeView}
        />
      )}
      {hasData && (
        <ExplorerLayout sidebarOpen={resourceList.opened}>
          <Box sx={{ display: "flex", alignItems: "center", gap: 2, px: 2, pt: 1 }}>
            <ToggleButtonGroup
              value={viewMode}
              exclusive
              onChange={(_, val) => {
                if (!val) return;
                if (val === "graph" && tableData) {
                  buildGraphFromFilteredData();
                }
                setViewMode(val);
              }}
              size="small"
            >
              <ToggleButton value="table">
                <Tooltip title="Table view">
                  <TableChartIcon fontSize="small" />
                </Tooltip>
              </ToggleButton>
              <ToggleButton value="graph" disabled={!tableData}>
                <Tooltip title="Graph view">
                  <AccountTreeIcon fontSize="small" />
                </Tooltip>
              </ToggleButton>
            </ToggleButtonGroup>
          </Box>

          {viewMode === "table" && tableData ? (
            <>
              <FilterBar
                filterState={filterState}
                onFilterChange={setFilterState}
                countsByType={tableData.countsByType}
                isUserTrait={tableData.isUserTrait}
                celError={celError}
                celLoading={celLoading}
                onCelApply={handleCelApply}
                onCelClear={handleCelClear}
              />
              <ResourceTable
                data={filteredTableData}
                isUserTrait={tableData.isUserTrait}
                hasMore={celResults === null && tableData.hasMore && !filterState.textSearch && !filterState.typeFilter && !filterState.entitlementSearch}
                loading={tableData.loading}
                onLoadMore={handleTableLoadMore}
                onRowClick={handleTableRowClick}
                countsByType={tableData.countsByType}
                totalCount={tableData.totalCount}
              />
            </>
          ) : viewMode === "graph" && hasGraphData ? (
            <TreeWrapper>
              <ReactFlow
                nodes={nodes}
                edges={edges}
                onNodesChange={onNodesChange}
                onEdgesChange={onEdgesChange}
                edgeTypes={edgeTypes}
                nodeTypes={nodeTypes}
                onEdgeClick={(e, edge) => {
                  if (edge.data?.entitlements?.[0]) {
                    handleEdgeClick(e, edge.data.entitlements[0]);
                  }
                }}
                onNodeClick={(e, node) =>
                  openResourceDetails(
                    node.data.resourceType,
                    extractPrincipalId(node.id),
                    node,
                    { top: e.clientY, left: e.clientX }
                  )
                }
                fitView
                attributionPosition="bottom-left"
              >
                <Controls position="bottom-right" showInteractive={false} />
              </ReactFlow>
            </TreeWrapper>
          ) : (
            <TreeWrapper />
          )}
        </ExplorerLayout>
      )}
      {resourceDetailsOpen.resource && (
        <ResourceDetailsModal
          resource={resourceDetailsOpen?.resource}
          closeDetails={closeResourceDetails}
          resourceDetails={resourceDetailsOpen}
          anchorPosition={detailsPosition}
        />
      )}
    </>
  );
};

export default Explorer;
