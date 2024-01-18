import React, { useContext, useEffect, useState } from "react";

const Context = React.createContext<ResourcesState>(null);

type Identities = {
  count: number;
  resourcesByType: any[];
  identityTypes: string[];
};

type ResourcesState = {
  resources: any;
  mappedResources: any;
  identities: Identities;
  groupTraitTypes: string[];
  roleTraitTypes: string[];
};

export const ResourcesContextProvider = ({ children }) => {
  const [resources, setResources] = useState<ResourcesState>({
    resources: {},
    mappedResources: {},
    identities: {
      count: 0,
      resourcesByType: [],
      identityTypes: [],
    },
    roleTraitTypes: [],
    groupTraitTypes: [],
  });

  useEffect(() => {
    const fetchData = async () => {
      const res = await (await fetch("/api/resources")).json();
      const map = {};
      const identityTypes = [];
      const resourcesByType = [];
      const groupTraitTypes = [];
      const roleTraitTypes = [];
      let resourcesCount = 0;

      await res.data.resources.forEach((resource) => {
        const type = resource.resource_type.id;
        if (!map[type]) {
          map[type] = [];
        }

        map[type].push(resource);
        map[type].sort((a, b) =>
          a.resource.display_name.localeCompare(b.resource.display_name)
        );

        switch (
          resource?.resource_type?.traits &&
          resource?.resource_type?.traits[0]
        ) {
          case 1:
            if (!resourcesByType[type]) {
              resourcesByType[type] = [];
            }
            resourcesCount += 1;
            resourcesByType[type].push(resource);
            break;
          case 2:
            if (!groupTraitTypes.includes(type)) {
              groupTraitTypes.push(type);
            }
            break;
          case 3:
            if (!roleTraitTypes.includes(type)) {
              roleTraitTypes.push(type);
            }
            break;
          default:
            break;
        }
      });

      setResources({
        resources: res.data,
        mappedResources: map,
        identities: {
          count: resourcesCount,
          resourcesByType: resourcesByType,
          identityTypes: identityTypes,
        },
        groupTraitTypes: groupTraitTypes,
        roleTraitTypes: roleTraitTypes,
      });
    };
    fetchData();
  }, []);

  return <Context.Provider value={resources}>{children}</Context.Provider>;
};

export const useResources = () => useContext(Context);
