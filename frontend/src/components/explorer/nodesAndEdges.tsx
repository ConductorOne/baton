import { Edge, Node } from "reactflow";
import dagre from "@dagrejs/dagre";
import { Position } from "reactflow";

const position = { x: 0, y: 0 };
const edgeType = "customEdge";
const nodeType = {
  parent: "parent",
  child: "child",
  expandable: "expandable",
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
  const resourceAccess = access.access && access.access;
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

  const users = [];
  let expandableGrantResourceId;

  resourceAccess &&
    resourceAccess.forEach((element, i) => {
      if (
        element.resource_type.traits &&
        element.resource_type.traits[0] === 1
      ) {
        users.push(element);
      } else {
        const expandableGrantType =
          "type.googleapis.com/c1.connector.v2.GrantExpandable";
        const isGroup =
          element.resource_type.traits && element.resource_type.traits[0] === 2;
        const isExpandable =
          isGroup &&
          element.grants[0].annotations &&
          element.grants[0].annotations[0].type_url === expandableGrantType;

        expandableGrantResourceId =
          isExpandable && element.resource.id.resource;
        nodes.push({
          id: isExpandable
            ? `expandable-${expandableGrantResourceId}`
            : `target-${element.resource.id.resource}`,
          data: {
            label: element.resource.display_name,
            targetHandle: `${access.resource.id.resource}-handle`,
            sourceHandle: `${element.resource.id.resource}-handle`,
            resourceType: element.resource_type.id,
            resourceTrait: element.resource_type.traits
              ? element.resource_type.traits[0]
              : 0,
          },
          position,
          type: isExpandable ? nodeType.expandable : nodeType.child,
        });

        edges.push({
          id: isExpandable
            ? `expandable-${access.resource.id.resource}-${element.resource.id.resource}`
            : `target-${access.resource.id.resource}-${element.resource.id.resource}`,
          source: `source-${access.resource.id.resource}`,
          target: isExpandable
            ? `expandable-${element.resource.id.resource}`
            : `target-${element.resource.id.resource}`,
          sourceHandle: `${access.resource.id.resource}-handle`,
          targetHandle: `${element.resource.id.resource}-handle`,
          label: "placeholder",
          type: edgeType,
          data: {
            entitlements: element.entitlements,
            openEntitlementsDetails: openEntitlementsDetails,
          },
        });
      }
    });

  users.length > 0 &&
    users.forEach((user) => {
      const sources = user.grants[0].sources && user.grants[0].sources.sources;
      let hasParent;
      for (let key in sources) {
        if (key.includes(expandableGrantResourceId)) {
          hasParent = true;
        }
      }

      const parent = hasParent
        ? `expandable-${expandableGrantResourceId}`
        : `source-${access.resource.id.resource}`;

      const parentHandle = hasParent
        ? `${expandableGrantResourceId}-handle`
        : `${access.resource.id.resource}-handle`;

      nodes.push({
        id: `target-${user.resource.id.resource}`,
        data: {
          label: user.resource.display_name,
          targetHandle: `${user.resource.id.resource}-handle`,
          resourceType: user.resource_type.id,
          sourceHandle: parentHandle,
          resourceTrait: user.resource_type.traits
            ? user.resource_type.traits[0]
            : 0,
        },
        position,
        type: nodeType.child,
      });

      edges.push({
        id: `child-${access.resource.id.resource}-${user.resource.id.resource}`,
        source: parent,
        target: `target-${user.resource.id.resource}`,
        sourceHandle: parentHandle,
        targetHandle: `${user.resource.id.resource}-handle`,
        label: "placeholder",
        type: edgeType,
        data: {
          entitlements: user.entitlements,
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
  const users = [];
  const otherResources = [];
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
      if (elem.resource_type.traits && elem.resource_type.traits[0] === 1) {
        users.push(elem);
      } else {
        otherResources.push(elem);
      }
    });

  otherResources.sort((a, b) =>
    a.resource_type.id.localeCompare(b.resource_type.id)
  );

  otherResources.forEach((resource) => {
    nodes.push({
      id: `target-${resource.resource.id.resource}`,
      data: {
        label: resource.resource.display_name,
        targetHandle: `${resource.resource.id.resource}-handle`,
        sourceHandle: `${principal.id.resource}-handle`,
        resourceType: resource.resource_type.id,
        resourceTrait: resource?.resource_type?.traits
          ? resource?.resource_type?.traits[0]
          : 0,
      },
      position,
      type: nodeType.child,
    });

    edges.push({
      id: `${principal.id.resource}-${resource.resource.id.resource}`,
      source: `source-${principal.id.resource}`,
      target: `target-${resource.resource.id.resource}`,
      sourceHandle: `${principal.id.resource}-handle`,
      targetHandle: `${resource.resource.id.resource}-handle`,
      label: "placeholder",
      type: edgeType,
      data: {
        entitlements: resource.entitlements,
        openEntitlementsDetails: openEntitlementsDetails,
      },
    });
  });

  users.forEach((user) => {
    nodes.push({
      id: `target-${user.resource.id.resource}`,
      data: {
        label: user.resource.display_name,
        targetHandle: `${user.resource.id.resource}-handle`,
        sourceHandle: `${principal.id.resource}-handle`,
        resourceType: user.resource_type.id,
        resourceTrait: user?.resource_type?.traits
          ? user?.resource_type?.traits[0]
          : 0,
      },
      position,
      type: nodeType.child,
    });

    edges.push({
      id: `${principal.id.resource}-${user.resource.id.resource}`,
      source: `source-${principal.id.resource}`,
      target: `target-${user.resource.id.resource}`,
      sourceHandle: `${principal.id.resource}-handle`,
      targetHandle: `${user.resource.id.resource}-handle`,
      label: "placeholder",
      type: edgeType,
      data: {
        entitlements: user.entitlements,
        openEntitlementsDetails: openEntitlementsDetails,
      },
    });
  });
  return { nodes, edges };
};
