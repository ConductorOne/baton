const capitalise = (str: string): string => {
  if (str === "") {
    return "";
  }
  return str.charAt(0).toUpperCase() + str.slice(1);
};

const removeUnderscore = (str: string): string => {
  if (str === "") {
    return "";
  }
  return str.replace(/_/g, ' ');
};

export const normalizeString = (str: string, shouldCapitalise: boolean): string => {
  let capitalized;

  if (shouldCapitalise) {
    capitalized = capitalise(str);
  } 

  return removeUnderscore(capitalized || str);
};

export const getResourceById = (resourceArray, resourceId: string) => {
  return resourceArray.find(obj => obj.resource.id.resource === resourceId) || null;
}

export const extractPrincipalId = (nodeId) => {
  const parts = nodeId.split(/source-|target-/);
  const principalId = parts[parts.length - 1];
  return principalId
}

export const isObjectEmpty = (obj) => {
  return Object.keys(obj).length === 0;
};
