import React, { useEffect, useState } from "react";
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
  ChildNode,
  ExpandableGrantNode,
  ParentNode,
} from "./components/customNode";
import { ResourcesSidebar } from "./components/resourcesSidebar";
import { useNavigate, useParams } from "react-router-dom";
import { TreeWrapper } from "./styles/styles";
import { ResourceDetailsModal } from "../resourceDetails";
import { extractPrincipalId, isObjectEmpty } from "../../common/helpers";
import {
  fetchAccessForUser,
  fetchGrantsForResource,
  fetchResourceDetails,
} from "./api";
import { populateNodes } from "./nodesAndEdges";
import { useTheme } from "@mui/material";
import { colors } from "../../style/colors";

type ResourceDetailsState = {
  resourceOpened?: boolean;
  entitlementOpened?: boolean;
  resource?: any;
};

const edgeTypes = { customEdge: CustomEdge };
const nodeTypes = {
  parent: ParentNode,
  child: ChildNode,
  expandable: ExpandableGrantNode,
};

const Explorer = ({ resourceList, closeResourceList }) => {
  const navigate = useNavigate();
  const theme = useTheme()
  const reactFlowInstance = useReactFlow();
  const nodesInitialized = useNodesInitialized();
  const { id: resourceId } = useParams();
  const { type: resourceType } = useParams();
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [resourceDetails, setResourceDetailsOpen] =
    useState<ResourceDetailsState>({
      resourceOpened: false,
      entitlementOpened: false,
    });

  useEffect(() => {
    if (resourceId) {
      const fetchData = async () => {
        const resource = await fetchResourceDetails(resourceType, resourceId);
        if (resource) {
          await openTreeView(resource);
        }
      };
      fetchData();
    }
  }, [resourceId, resourceType]);

  useEffect(() => {
    if (nodesInitialized) {
      reactFlowInstance.fitView();
    }
  }, [nodesInitialized]);

  const openEntitlementsDetails = (entitlement) => {
    setResourceDetailsOpen({
      entitlementOpened: true,
      resource: entitlement,
    });
  };

  const openResourceDetails = async (
    resourceType: string,
    principalId: string,
    currentNode
  ) => {
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

    edges &&
      edges.forEach((edge) => {
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
              theme.palette.mode === "light" ? colors.batonGreen600 : colors.batonGreen500
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
  };

  const openTreeView = async (resource) => {
    navigate(`/${resource.resource_type.id}/${resource.resource.id.resource}`);
    let grantsAccess;
    if (
      resource.resource_type.traits &&
      // user trait for access
      resource.resource_type.traits[0] === 1
    ) {
      grantsAccess = await fetchAccessForUser(
        resource.resource_type.id,
        resource.resource.id.resource
      );
    } else {
      grantsAccess = await fetchGrantsForResource(
        resource.resource_type.id,
        resource.resource.id.resource
      );
    }
    if (!grantsAccess.access || !isObjectEmpty(grantsAccess.access)) {
      populateNodes(
        setNodes,
        setEdges,
        grantsAccess,
        resource,
        openEntitlementsDetails
      );
    }
    closeResourceDetails();
  };

  const handleEdgeClick = (e, entitlement) => {
    e.stopPropagation();
    openEntitlementsDetails(entitlement);
  };

  return (
    <>
      {resourceList.opened && (
        <ResourcesSidebar
          closeResourceList={closeResourceList}
          resourceType={resourceList.resource}
          openTreeView={openTreeView}
        />
      )}
      {nodes.length > 0 && edges.length === nodes.length - 1 && (
        <TreeWrapper>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            edgeTypes={edgeTypes}
            nodeTypes={nodeTypes}
            onEdgeClick={(e, edge) =>
              handleEdgeClick(e, edge.data.entitlements[0])
            }
            onNodeClick={(e, node) =>
              openResourceDetails(
                node.data.resourceType,
                extractPrincipalId(node.id),
                node
              )
            }
            fitView
            attributionPosition="bottom-left"
          >
            <Controls position="bottom-right" showInteractive={false} />
          </ReactFlow>
        </TreeWrapper>
      )}
      {resourceDetails.resource && (
        <ResourceDetailsModal
          resource={resourceDetails?.resource}
          closeDetails={closeResourceDetails}
          resourceDetails={resourceDetails}
        />
      )}
    </>
  );
};

export default Explorer;
