import React, { useState } from "react";
import Footer from "../components/footer/footer";
import { Navigation } from "../components/navigation/navigation";
import Explorer from "../components/explorer/explorer";
import { ReactFlowProvider } from "reactflow";

type ResourcesListState = {
  opened: boolean;
  resource?: string;
};

const Homepage = () => {
  const [resourceList, setResourcesList] = useState<ResourcesListState>({
    opened: false,
  });

  const openResourceList = (resourceType: string) => {
    setResourcesList({
      opened: true,
      resource: resourceType,
    });
  };

  const closeResourceList = () => {
    setResourcesList({
      opened: false,
    });
  };

  return (
    <>
      <Navigation openResourceList={openResourceList} resourceState={resourceList}/>
      <ReactFlowProvider>
        <Explorer
          resourceList={resourceList}
          closeResourceList={closeResourceList}
        />
      </ReactFlowProvider>
      <Footer />
    </>
  );
};

export default Homepage;
