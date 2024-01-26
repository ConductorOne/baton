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

const isCamelCase = (str: string) => /^[a-z]+(?:[A-Z][a-z]*)*$/.test(str);

const separateCamelCase = (str: string) => str.replace(/([A-Z])/g, ' $1').trim();

export const normalizeString = (str: string, shouldCapitalise: boolean): string => {
  let capitalized;
  let string = str;

  if (isCamelCase(str)) {
    string = separateCamelCase(str);
  }

  if (shouldCapitalise) {
    capitalized = capitalise(string);
  }

  return removeUnderscore(capitalized || string);
};

export const getResourceById = (resourceArray, resourceId: string) => {
  return resourceArray.find(obj => obj.resource.id.resource === resourceId) || null;
};

export const extractPrincipalId = (nodeId) => {
  const parts = nodeId.split(/source-|target-|expandable-/);
  const principalId = parts[parts.length - 1];
  return principalId;
};

export const isObjectEmpty = (obj) => {
  return Object.keys(obj).length === 0;
};
