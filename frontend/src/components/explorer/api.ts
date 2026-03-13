export interface PaginatedResponse<T = any> {
  data: T;
  next_page_token?: string;
  total_count?: number;
  counts_by_type?: Record<string, number>;
}

async function fetchData(url: string) {
  const resp = await fetch(url);
  if (!resp.ok) {
    throw new Error(`Request failed: ${resp.status} ${resp.statusText}`);
  }
  const json = await resp.json();
  return json.data;
}

async function fetchPaginated<T = any>(url: string): Promise<PaginatedResponse<T>> {
  const resp = await fetch(url);
  if (!resp.ok) {
    throw new Error(`Request failed: ${resp.status} ${resp.statusText}`);
  }
  return resp.json();
}

export const fetchAccessForUser = async (resourceType: string, resource: string) => {
  return fetchData(`/api/access/${resourceType}/${resource}`);
};

export const fetchGrantsForResource = async (
  resourceType: string,
  resource: string,
  pageSize?: number,
  pageToken?: string,
): Promise<PaginatedResponse> => {
  const params = new URLSearchParams();
  if (pageSize) params.set("page_size", String(pageSize));
  if (pageToken) params.set("page_token", pageToken);
  const qs = params.toString();
  return fetchPaginated(`/api/grants/${resourceType}/${resource}${qs ? `?${qs}` : ""}`);
};

export const fetchResourceDetails = async (resourceType: string, resource: string) => {
  return fetchData(`/api/${resourceType}/${resource}`);
};

export const fetchResourceTypes = async () => {
  return fetchData("/api/resourceTypes");
};

export const fetchResourcesWithUserCount = async (resourceType: string) => {
  return fetchData(`/api/principals/${resourceType}`);
};

export const searchWithCEL = async (params: {
  cel: string;
  scope: "resources" | "grants";
  resourceTypeId?: string;
  resourceType?: string;
  resourceId?: string;
  pageSize?: number;
  pageToken?: string;
}): Promise<PaginatedResponse> => {
  const qs = new URLSearchParams();
  qs.set("cel", params.cel);
  qs.set("scope", params.scope);
  if (params.resourceTypeId) qs.set("resource_type_id", params.resourceTypeId);
  if (params.resourceType) qs.set("resource_type", params.resourceType);
  if (params.resourceId) qs.set("resource_id", params.resourceId);
  if (params.pageSize) qs.set("page_size", String(params.pageSize));
  if (params.pageToken) qs.set("page_token", params.pageToken);
  return fetchPaginated(`/api/search?${qs.toString()}`);
};

export const fetchResourcesByType = async (
  resourceTypeId: string,
  pageToken?: string,
): Promise<PaginatedResponse> => {
  const params = new URLSearchParams({ resource_type_id: resourceTypeId });
  if (pageToken) params.set("page_token", pageToken);
  return fetchPaginated(`/api/resources?${params.toString()}`);
};
