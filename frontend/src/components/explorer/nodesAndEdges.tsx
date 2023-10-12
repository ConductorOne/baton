import { Edge, Node } from "reactflow";
import dagre from "@dagrejs/dagre";
import { Position } from "reactflow";

const position = { x: 0, y: 0 };
const edgeType = "customEdge";
const nodeType = {
  parent: "parent",
  child: "child",
};

const EdgeStyle = {
  stroke: "rgba(115, 189, 81, 1)",
};

const createGraphLayout = (nodes, edges) => {
  const g = new dagre.graphlib.Graph();
  g.setGraph({ rankdir: "LR" });

  g.setDefaultEdgeLabel(() => ({}));
  const nodeWidth = 450;
  const nodeHeight = 80;
  nodes.forEach((node) => {
    g.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach((edge) => {
    g.setEdge(edge.source, edge.target);
  });

  dagre.layout(g);

  nodes.forEach((node) => {
    const nodeWithPosition = g.node(node.id);
    node.targetPosition = Position.Left;
    node.sourcePosition = Position.Right;
    node.selectable = true;
    node.focusable = false;

    node.position = {
      x: nodeWithPosition.x - nodeWidth / 2,
      y: nodeWithPosition.y - nodeHeight / 2,
    };

    return node;
  });

  return { nodes, edges };
};

export const populateNodes = (
  setNodes,
  setEdges,
  access,
  resource,
  openEntitlementsDetails
) => {
  if (
    resource.resource_type.traits &&
    // user trait is 1
    resource.resource_type.traits[0] === 1
  ) {
    const { nodes: initialNodes, edges: initialEdges } =
      populateNodesAndEdgesForPrincipals(access, openEntitlementsDetails);

    const {
      nodes: layoutedNodesForPrincipals,
      edges: layoutedEdgesForPrincipals,
    } = createGraphLayout(initialNodes, initialEdges);

    setNodes(layoutedNodesForPrincipals);
    setEdges(layoutedEdgesForPrincipals);
  } else {
    const { nodes: n, edges: e } = populateNodesAndEdgesForGrants(
      access,
      openEntitlementsDetails
    );

    const { nodes: layoutedNodes, edges: layoutedEdges } = createGraphLayout(
      n,
      e
    );
    setNodes(layoutedNodes);
    setEdges(layoutedEdges);
  }
};

export const populateNodesAndEdgesForGrants = (
  access,
  openEntitlementsDetails
  ) => {
  const edges: Edge[] = [];
  const nodes: Node[] = [
    {
      id: `source-${access.resource.id.resource}`,
      data: {
        label: access.resource.display_name,
        sourceHandle: `${access.resource.id.resource}-handle`,
        resourceTrait: access.resource_type?.traits
          ? access.resource_type.traits[0]
          : 0,
        resourceType: access.resource_type.id,
      },
      position,
      type: nodeType.parent,
    },
  ];

  access.access &&
    access.access.forEach((element, i) => {
      nodes.push({
        id: `target-${element.resource.id.resource}`,
        data: {
          label: element.resource.display_name,
          targetHandle: `${element.resource.id.resource}-handle`,
          resourceType: element.resource_type.id,
          resourceTrait: element.resource_type.traits
            ? element.resource_type.traits[0]
            : 0,
        },
        position,
        type: nodeType.child,
      });

      edges.push({
        id: `${access.resource.id.resource}-${element.resource.id.resource}`,
        source: `source-${access.resource.id.resource}`,
        target: `target-${element.resource.id.resource}`,
        sourceHandle: `${access.resource.id.resource}-handle`,
        targetHandle: `${element.resource.id.resource}-handle`,
        label: "placeholder",
        type: edgeType,
        style: EdgeStyle,
        data: {
          entitlements: element.entitlements,
          openEntitlementsDetails: openEntitlementsDetails,
        },
      });
    });
  return { nodes, edges };
};

export const populateNodesAndEdgesForPrincipals = (
  userAccess,
  openEntitlementsDetails
) => {
  const principal = userAccess?.principal;
  const access = userAccess?.access;
  const edges: Edge[] = [];
  const nodes: Node[] = [
    {
      id: `source-${principal.id.resource}`,
      data: {
        label: principal.display_name,
        resourceTrait: 1,
        resourceType: principal.id.resource_type,
        sourceHandle: `${principal.id.resource}-handle`,
      },
      position,
      type: nodeType.parent,
    },
  ];

  access &&
    access.forEach((elem) => {
      nodes.push({
        id: `target-${elem.resource.id.resource}`,
        data: {
          label: elem.resource.display_name,
          targetHandle: `${elem.resource.id.resource}-handle`,
          resourceType: elem.resource_type.id,
          resourceTrait: elem?.resource_type?.traits
            ? elem?.resource_type?.traits[0]
            : 0,
        },
        position,
        type: nodeType.child,
      });

      edges.push({
        id: `${principal.id.resource}-${elem.resource.id.resource}`,
        source: `source-${principal.id.resource}`,
        target: `target-${elem.resource.id.resource}`,
        sourceHandle: `${principal.id.resource}-handle`,
        targetHandle: `${elem.resource.id.resource}-handle`,
        label: "placeholder",
        type: edgeType,
        style: EdgeStyle,
        data: {
          entitlements: elem.entitlements,
          openEntitlementsDetails: openEntitlementsDetails,
        },
      });
    });
  return { nodes, edges };
};
