import { Edge, Node } from "reactflow";
import { Position } from "reactflow";
import pluralize from "pluralize";
import { normalizeString } from "../../common/helpers";

const position = { x: 0, y: 0 };
const edgeType = "customEdge";
const nodeType = {
  parent: "parent",
  child: "child",
  expandable: "expandable",
  aggregate: "aggregate",
  entitlement: "entitlement",
};

export const AGGREGATION_THRESHOLD = 100;

const createGraphLayout = (nodes, edges) => {
  const nodeWidth = 300;
  const nodeHeight = 80;
  const horizontalGap = 200;
  const verticalGap = 20;

  // Group nodes by layer (default: parent=0, others=1)
  const layers: Map<number, any[]> = new Map();
  for (const node of nodes) {
    const layer = node.data.layer ?? (node.type === nodeType.parent ? 0 : 1);
    if (!layers.has(layer)) layers.set(layer, []);
    layers.get(layer)!.push(node);
  }

  const sortedLayers = [...layers.keys()].sort((a, b) => a - b);

  for (const layerIdx of sortedLayers) {
    const layerNodes = layers.get(layerIdx)!;
    const x = layerIdx * (nodeWidth + horizontalGap);
    const totalHeight = layerNodes.length * (nodeHeight + verticalGap) - verticalGap;
    const startY = -totalHeight / 2;

    for (let i = 0; i < layerNodes.length; i++) {
      layerNodes[i].position = { x, y: startY + i * (nodeHeight + verticalGap) };
      layerNodes[i].targetPosition = Position.Left;
      layerNodes[i].sourcePosition = Position.Right;
      layerNodes[i].selectable = true;
      layerNodes[i].focusable = false;
    }
  }

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
      populateNodesAndEdgesForPrincipals(access);

    const {
      nodes: layoutedNodesForPrincipals,
      edges: layoutedEdgesForPrincipals,
    } = createGraphLayout(initialNodes, initialEdges);

    setNodes(layoutedNodesForPrincipals);
    setEdges(layoutedEdgesForPrincipals);
  } else {
    const { nodes: n, edges: e } = populateNodesAndEdgesForGrants(access);

    const { nodes: layoutedNodes, edges: layoutedEdges } = createGraphLayout(
      n,
      e
    );
    setNodes(layoutedNodes);
    setEdges(layoutedEdges);
  }
};

// Grants view: Source Resource → Entitlement Nodes → Principal/User Nodes
export const populateNodesAndEdgesForGrants = (access) => {
  const resourceAccess = access.access || [];
  const edges: Edge[] = [];
  const sourceId = access.resource.id.resource;

  const nodes: Node[] = [
    {
      id: `source-${sourceId}`,
      data: {
        label: access.resource.display_name,
        sourceHandle: `${sourceId}-handle`,
        resourceTrait: access.resource_type?.traits
          ? access.resource_type.traits[0]
          : 0,
        resourceType: access.resource_type.id,
        layer: 0,
      },
      position,
      type: nodeType.parent,
    },
  ];

  // Collect unique entitlements by slug and map entitlement → principals
  const entitlementMap = new Map<string, { entitlement: any; principals: Set<string> }>();
  const principalMap = new Map<string, any>();

  for (const item of resourceAccess) {
    const principalId = item.resource.id.resource;
    if (!principalMap.has(principalId)) {
      principalMap.set(principalId, item);
    }

    for (const ent of (item.entitlements || [])) {
      const entKey = ent.slug || ent.display_name || ent.id;
      if (!entitlementMap.has(entKey)) {
        entitlementMap.set(entKey, { entitlement: ent, principals: new Set() });
      }
      entitlementMap.get(entKey)!.principals.add(principalId);
    }
  }

  // Create entitlement nodes (layer 1)
  for (const [entKey, { entitlement }] of entitlementMap) {
    nodes.push({
      id: `entitlement-${entKey}`,
      data: {
        label: entitlement.display_name || entitlement.slug,
        targetHandle: `ent-${entKey}-target`,
        sourceHandle: `ent-${entKey}-source`,
        isEntitlement: true,
        entitlement,
        layer: 1,
      },
      position,
      type: nodeType.entitlement,
    });

    // Edge from source to entitlement
    edges.push({
      id: `source-ent-${sourceId}-${entKey}`,
      source: `source-${sourceId}`,
      target: `entitlement-${entKey}`,
      sourceHandle: `${sourceId}-handle`,
      targetHandle: `ent-${entKey}-target`,
      type: edgeType,
      data: {},
    });
  }

  // Create principal/user nodes (layer 2)
  for (const [principalId, item] of principalMap) {
    nodes.push({
      id: `target-${principalId}`,
      data: {
        label: item.resource.display_name,
        targetHandle: `${principalId}-handle`,
        resourceType: item.resource_type.id,
        resourceTrait: item.resource_type?.traits?.[0] || 0,
        layer: 2,
      },
      position,
      type: nodeType.child,
    });
  }

  // Edges from entitlements to principals
  for (const [entKey, { principals }] of entitlementMap) {
    for (const principalId of principals) {
      edges.push({
        id: `ent-principal-${entKey}-${principalId}`,
        source: `entitlement-${entKey}`,
        target: `target-${principalId}`,
        sourceHandle: `ent-${entKey}-source`,
        targetHandle: `${principalId}-handle`,
        type: edgeType,
        data: {},
      });
    }
  }

  return { nodes, edges };
};

// Principals view: User → Entitlement Nodes → Resource Nodes
export const populateNodesAndEdgesForPrincipals = (userAccess) => {
  const principal = userAccess?.principal;
  const access = userAccess?.access || [];
  const edges: Edge[] = [];
  const sourceId = principal.id.resource;

  const nodes: Node[] = [
    {
      id: `source-${sourceId}`,
      data: {
        label: principal.display_name,
        resourceTrait: 1,
        resourceType: principal.id.resource_type,
        sourceHandle: `${sourceId}-handle`,
        layer: 0,
      },
      position,
      type: nodeType.parent,
    },
  ];

  // Collect unique entitlements by slug and map entitlement → target resources
  const entitlementMap = new Map<string, { entitlement: any; resourceIds: Set<string> }>();
  const resourceMap = new Map<string, any>();

  for (const item of access) {
    const resId = item.resource.id.resource;
    if (!resourceMap.has(resId)) {
      resourceMap.set(resId, item);
    }

    for (const ent of (item.entitlements || [])) {
      const entKey = ent.slug || ent.display_name || ent.id;
      if (!entitlementMap.has(entKey)) {
        entitlementMap.set(entKey, { entitlement: ent, resourceIds: new Set() });
      }
      entitlementMap.get(entKey)!.resourceIds.add(resId);
    }
  }

  // Create entitlement nodes (layer 1)
  for (const [entKey, { entitlement }] of entitlementMap) {
    nodes.push({
      id: `entitlement-${entKey}`,
      data: {
        label: entitlement.display_name || entitlement.slug,
        targetHandle: `ent-${entKey}-target`,
        sourceHandle: `ent-${entKey}-source`,
        isEntitlement: true,
        entitlement,
        layer: 1,
      },
      position,
      type: nodeType.entitlement,
    });

    // Edge from source to entitlement
    edges.push({
      id: `source-ent-${sourceId}-${entKey}`,
      source: `source-${sourceId}`,
      target: `entitlement-${entKey}`,
      sourceHandle: `${sourceId}-handle`,
      targetHandle: `ent-${entKey}-target`,
      type: edgeType,
      data: {},
    });
  }

  // Create resource nodes (layer 2)
  for (const [resId, item] of resourceMap) {
    nodes.push({
      id: `target-${resId}`,
      data: {
        label: item.resource.display_name,
        targetHandle: `${resId}-handle`,
        resourceType: item.resource_type.id,
        resourceTrait: item.resource_type?.traits?.[0] || 0,
        layer: 2,
      },
      position,
      type: nodeType.child,
    });
  }

  // Edges from entitlements to resources
  for (const [entKey, { resourceIds }] of entitlementMap) {
    for (const resId of resourceIds) {
      edges.push({
        id: `ent-res-${entKey}-${resId}`,
        source: `entitlement-${entKey}`,
        target: `target-${resId}`,
        sourceHandle: `ent-${entKey}-source`,
        targetHandle: `${resId}-handle`,
        type: edgeType,
        data: {},
      });
    }
  }

  return { nodes, edges };
};

const MAX_FEATURED_NODES = 5;

// Creates a summary graph with aggregate type nodes and a few featured real member nodes.
export const populateSummaryNodes = (
  setNodes,
  setEdges,
  access,
  resource,
  countsByType: Record<string, number>,
  firstPageAccess: any[],
  openEntitlementsDetails,
) => {
  const edges: Edge[] = [];
  const sourceId = access.resource.id.resource;
  const nodes: Node[] = [
    {
      id: `source-${sourceId}`,
      data: {
        label: access.resource.display_name,
        sourceHandle: `${sourceId}-handle`,
        resourceTrait: access.resource_type?.traits
          ? access.resource_type.traits[0]
          : 0,
        resourceType: access.resource_type.id,
        layer: 0,
      },
      position,
      type: nodeType.parent,
    },
  ];

  // Collect entitlements and traits from the first page of access data.
  const entitlementsByType: Record<string, any[]> = {};
  const traitsByType: Record<string, number> = {};
  const accessItems = access.access || [];
  for (const item of accessItems) {
    const typeId = item.resource_type?.id;
    if (typeId) {
      if (!entitlementsByType[typeId]) {
        entitlementsByType[typeId] = item.entitlements || [];
      }
      if (traitsByType[typeId] === undefined) {
        traitsByType[typeId] = item.resource_type?.traits?.[0] || 0;
      }
    }
  }

  // Aggregate nodes per type
  for (const [typeId, count] of Object.entries(countsByType)) {
    const label = `${count} ${pluralize(normalizeString(typeId, true))}`;
    const trait = traitsByType[typeId] || 0;
    nodes.push({
      id: `aggregate-${typeId}`,
      data: {
        label,
        targetHandle: `aggregate-${typeId}-handle`,
        sourceHandle: `${sourceId}-handle`,
        resourceType: typeId,
        resourceTrait: trait,
        isAggregate: true,
        aggregateType: typeId,
        aggregateCount: count,
        layer: 1,
      },
      position,
      type: nodeType.aggregate,
    });

    edges.push({
      id: `agg-${sourceId}-${typeId}`,
      source: `source-${sourceId}`,
      target: `aggregate-${typeId}`,
      sourceHandle: `${sourceId}-handle`,
      targetHandle: `aggregate-${typeId}-handle`,
      type: edgeType,
      data: {
        entitlements: entitlementsByType[typeId] || [],
        openEntitlementsDetails,
      },
    });
  }

  // Featured nodes: up to 5 real members from the first page
  const featured = firstPageAccess.slice(0, MAX_FEATURED_NODES);
  for (const item of featured) {
    const resId = item.resource?.id?.resource;
    if (!resId) continue;
    const trait = item.resource_type?.traits?.[0] || 0;
    nodes.push({
      id: `target-${resId}`,
      data: {
        label: item.resource.display_name,
        targetHandle: `${resId}-handle`,
        sourceHandle: `${sourceId}-handle`,
        resourceType: item.resource_type?.id || "",
        resourceTrait: trait,
        layer: 1,
      },
      position,
      type: nodeType.child,
    });

    edges.push({
      id: `featured-${sourceId}-${resId}`,
      source: `source-${sourceId}`,
      target: `target-${resId}`,
      sourceHandle: `${sourceId}-handle`,
      targetHandle: `${resId}-handle`,
      type: edgeType,
      data: {
        entitlements: item.entitlements || [],
        openEntitlementsDetails,
      },
    });
  }

  const { nodes: layoutedNodes, edges: layoutedEdges } = createGraphLayout(nodes, edges);
  setNodes(layoutedNodes);
  setEdges(layoutedEdges);
};

// Creates a drilldown graph showing individual members of a single type with entitlement nodes.
export const populateDrilldownNodes = (
  setNodes,
  setEdges,
  parentResource,
  members: any[],
  openEntitlementsDetails,
) => {
  const sourceId = parentResource.resource.id.resource;
  const edges: Edge[] = [];
  const nodes: Node[] = [
    {
      id: `source-${sourceId}`,
      data: {
        label: parentResource.resource.display_name,
        sourceHandle: `${sourceId}-handle`,
        resourceTrait: parentResource.resource_type?.traits
          ? parentResource.resource_type.traits[0]
          : 0,
        resourceType: parentResource.resource_type.id,
        layer: 0,
      },
      position,
      type: nodeType.parent,
    },
  ];

  // Build entitlement → member mapping
  const entitlementMap = new Map<string, { entitlement: any; memberIds: Set<string> }>();
  const memberMap = new Map<string, any>();

  for (const item of members) {
    const resId = item.resource?.id?.resource;
    if (!resId) continue;
    if (!memberMap.has(resId)) {
      memberMap.set(resId, item);
    }

    for (const ent of (item.entitlements || [])) {
      const entKey = ent.slug || ent.display_name || ent.id;
      if (!entitlementMap.has(entKey)) {
        entitlementMap.set(entKey, { entitlement: ent, memberIds: new Set() });
      }
      entitlementMap.get(entKey)!.memberIds.add(resId);
    }
  }

  // Entitlement nodes (layer 1)
  for (const [entId, { entitlement }] of entitlementMap) {
    nodes.push({
      id: `entitlement-${entId}`,
      data: {
        label: entitlement.display_name || entitlement.slug,
        targetHandle: `ent-${entId}-target`,
        sourceHandle: `ent-${entId}-source`,
        isEntitlement: true,
        entitlement,
        layer: 1,
      },
      position,
      type: nodeType.entitlement,
    });

    edges.push({
      id: `source-ent-${sourceId}-${entId}`,
      source: `source-${sourceId}`,
      target: `entitlement-${entId}`,
      sourceHandle: `${sourceId}-handle`,
      targetHandle: `ent-${entId}-target`,
      type: edgeType,
      data: {},
    });
  }

  // Member nodes (layer 2)
  for (const [resId, item] of memberMap) {
    const trait = item.resource_type?.traits?.[0] || 0;
    nodes.push({
      id: `target-${resId}`,
      data: {
        label: item.resource.display_name,
        targetHandle: `${resId}-handle`,
        resourceType: item.resource_type?.id || "",
        resourceTrait: trait,
        layer: 2,
      },
      position,
      type: nodeType.child,
    });
  }

  // Edges from entitlements to members
  for (const [entId, { memberIds }] of entitlementMap) {
    for (const resId of memberIds) {
      edges.push({
        id: `ent-member-${entId}-${resId}`,
        source: `entitlement-${entId}`,
        target: `target-${resId}`,
        sourceHandle: `ent-${entId}-source`,
        targetHandle: `${resId}-handle`,
        type: edgeType,
        data: {},
      });
    }
  }

  const { nodes: layoutedNodes, edges: layoutedEdges } = createGraphLayout(nodes, edges);
  setNodes(layoutedNodes);
  setEdges(layoutedEdges);
};
