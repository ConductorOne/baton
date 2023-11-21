import React from "react";
import Explorer from "../components/explorer/explorer";
import { ReactFlowProvider } from "reactflow";

export const ExplorerPage = ({ resourceList, closeResourceList}) => {
  return (
    <ReactFlowProvider>
      <Explorer
        resourceList={resourceList}
        closeResourceList={closeResourceList}
      />
    </ReactFlowProvider>
  );
};
