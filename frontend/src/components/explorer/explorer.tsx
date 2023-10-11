import React, { useEffect, useState } from "react";
import ReactFlow, {
  Controls,
  useEdgesState,
  useNodesInitialized,
  useNodesState,
  useReactFlow,
} from "reactflow";
import "reactflow/dist/style.css";
import { CustomEdge } from "./components/customEdge";
import { ChildNode, ParentNode } from "./components/customNode";
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

type ResourceDetailsState = {
  resourceOpened?: boolean;
  entitlementOpened?: boolean;
  resource?: any;
};

const edgeTypes = { customEdge: CustomEdge };
const nodeTypes = {
  parent: ParentNode,
  child: ChildNode,
};

const Explorer = ({ resourceList, closeResourceList }) => {
  const navigate = useNavigate();
  const reactFlowInstance = useReactFlow();
  const nodesInitialized = useNodesInitialized();
  const { id: resourceId } = useParams();
  const { type: resourceType } = useParams();
  const [nodes, setNodes] = useNodesState([]);
  const [edges, setEdges] = useEdgesState([]);
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

  const openResourceDetails = async (resourceType: string, nodeId: string) => {
    const resourceDetails = await fetchResourceDetails(resourceType, nodeId);
    setResourceDetailsOpen({
      resourceOpened: true,
      resource: resourceDetails,
    });
  };

  const closeResourceDetails = () => {
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
            edgeTypes={edgeTypes}
            nodeTypes={nodeTypes}
            onEdgeClick={(e, edge) =>
              handleEdgeClick(e, edge.data.entitlements[0])
            }
            onNodeClick={(e, node) =>
              openResourceDetails(
                node.data.resourceType,
                extractPrincipalId(node.id)
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
