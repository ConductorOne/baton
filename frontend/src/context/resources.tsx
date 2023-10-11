import React, { useContext, useEffect, useState } from "react";

const Context = React.createContext<ResourcesState>(null);

type ResourcesState = {
  resources: any;
  mappedResources: any;
};

export const ResourcesContextProvider = ({ children }) => {
  const [resources, setResources] = useState<ResourcesState>({
    resources: {},
    mappedResources: {},
  });

  useEffect(() => {
    const fetchData = async () => {
      const res = await (await fetch("/api/resources")).json();
      const map = {};
      res.data.resources.forEach((resource) => {
        const type = resource.resource_type.id;
        if (!map[type]) {
          map[type] = [];
        }

        map[type].push(resource);
        map[type].sort((a, b) => a.resource.display_name.localeCompare(b.resource.display_name))
      });

      setResources({
        resources: res.data,
        mappedResources: map,
      });
    };
    fetchData();
  }, []);

  return <Context.Provider value={resources}>{children}</Context.Provider>;
};

export const useResources = () => useContext(Context);
