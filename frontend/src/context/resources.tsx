import React, { useCallback, useContext, useEffect, useState } from "react";
import { fetchResourcesByType as apiFetchResourcesByType, PaginatedResponse } from "../components/explorer/api";

const Context = React.createContext<ResourcesState>(null);

type Identities = {
  count: number;
  resourcesByType: any[];
  identityTypes: string[];
};

type PaginatedResourceCache = {
  resources: any[];
  nextPageToken?: string;
  totalCount?: number;
};

type ResourcesState = {
  resources: any;
  mappedResources: any;
  identities: Identities;
  groupTraitTypes: string[];
  roleTraitTypes: string[];
  loading: boolean;
  error: string | null;
  fetchResourcesByType: (resourceTypeId: string) => Promise<any[]>;
  fetchResourcePage: (resourceTypeId: string, pageToken?: string) => Promise<PaginatedResponse>;
  getResourceCache: (resourceTypeId: string) => PaginatedResourceCache | undefined;
  appendResources: (resourceTypeId: string, newResources: any[], nextPageToken?: string) => void;
};

export const ResourcesContextProvider = ({ children }) => {
  const [state, setState] = useState<
    Omit<ResourcesState, "fetchResourcesByType" | "fetchResourcePage" | "getResourceCache" | "appendResources">
  >({
    resources: {},
    mappedResources: {},
    identities: {
      count: 0,
      resourcesByType: [],
      identityTypes: [],
    },
    roleTraitTypes: [],
    groupTraitTypes: [],
    loading: true,
    error: null,
  });

  const [resourceCache, setResourceCache] = useState<Record<string, PaginatedResourceCache>>({});

  const fetchResourcesByType = useCallback(async (resourceTypeId: string): Promise<any[]> => {
    if (resourceCache[resourceTypeId]) {
      return resourceCache[resourceTypeId].resources;
    }

    const resp = await apiFetchResourcesByType(resourceTypeId);
    const resources = resp.data?.resources || [];

    setResourceCache((prev) => ({
      ...prev,
      [resourceTypeId]: {
        resources,
        nextPageToken: resp.next_page_token,
        totalCount: resp.total_count,
      },
    }));
    return resources;
  }, [resourceCache]);

  const fetchResourcePage = useCallback(async (resourceTypeId: string, pageToken?: string): Promise<PaginatedResponse> => {
    return apiFetchResourcesByType(resourceTypeId, pageToken);
  }, []);

  const getResourceCache = useCallback((resourceTypeId: string): PaginatedResourceCache | undefined => {
    return resourceCache[resourceTypeId];
  }, [resourceCache]);

  const appendResources = useCallback((resourceTypeId: string, newResources: any[], nextPageToken?: string) => {
    setResourceCache((prev) => {
      const existing = prev[resourceTypeId];
      return {
        ...prev,
        [resourceTypeId]: {
          resources: [...(existing?.resources || []), ...newResources],
          nextPageToken,
          totalCount: existing?.totalCount,
        },
      };
    });
  }, []);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const resp = await fetch("/api/resourceTypes");
        if (!resp.ok) {
          throw new Error("Failed to fetch resource types");
        }
        const res = await resp.json();

        const identityTypes: string[] = [];
        const groupTraitTypes: string[] = [];
        const roleTraitTypes: string[] = [];

        for (const rtOutput of res.data?.resource_types || []) {
          const rt = rtOutput.resource_type;
          if (!rt?.traits?.length) continue;

          switch (rt.traits[0]) {
            case 1: // TRAIT_USER
              identityTypes.push(rt.id);
              break;
            case 2: // TRAIT_GROUP
              groupTraitTypes.push(rt.id);
              break;
            case 3: // TRAIT_ROLE
              roleTraitTypes.push(rt.id);
              break;
            default:
              break;
          }
        }

        // Fetch first page of identity resources to get counts.
        let identityCount = 0;
        const resourcesByType: any[] = [];
        const mappedResources: Record<string, any[]> = {};
        const newResourceCache: Record<string, PaginatedResourceCache> = {};

        for (const typeId of identityTypes) {
          const typeResp = await apiFetchResourcesByType(typeId);
          const resources = typeResp.data?.resources || [];

          mappedResources[typeId] = resources;
          resourcesByType[typeId] = resources;
          newResourceCache[typeId] = {
            resources,
            nextPageToken: typeResp.next_page_token,
            totalCount: typeResp.total_count,
          };
          identityCount += typeResp.total_count || resources.length;
        }

        setResourceCache(newResourceCache);

        setState({
          resources: { resource_types: res.data?.resource_types },
          mappedResources,
          identities: {
            count: identityCount,
            resourcesByType,
            identityTypes,
          },
          groupTraitTypes,
          roleTraitTypes,
          loading: false,
          error: null,
        });
      } catch (err) {
        setState((prev) => ({
          ...prev,
          loading: false,
          error: err instanceof Error ? err.message : "Unknown error",
        }));
      }
    };
    fetchData();
  }, []);

  const value: ResourcesState = {
    ...state,
    fetchResourcesByType,
    fetchResourcePage,
    getResourceCache,
    appendResources,
  };

  return <Context.Provider value={value}>{children}</Context.Provider>;
};

export const useResources = () => useContext(Context);
