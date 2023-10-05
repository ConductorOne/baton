import React, { useState } from "react";
import Footer from "../components/footer/footer";
import "../components/explorer/index.css";
import { Navigation } from "../components/navigation/navigation";
import Explorer from "../components/explorer/explorer";

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
      <Navigation openResourceList={openResourceList} />
      <Explorer
        resourceList={resourceList}
        closeResourceList={closeResourceList}
      />
      <Footer />
    </>
  );
};

export default Homepage;
