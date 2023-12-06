export const fetchAccessForUser = async (resourceType, resource) => {
  const access = await (
    await fetch(`/api/access/${resourceType}/${resource}`)
  ).json();
  return access.data;
};

export const fetchGrantsForResource = async (resourceType, resource) => {
  const access = await (
    await fetch(`/api/grants/${resourceType}/${resource}`)
  ).json();
  return access.data;
};

export const fetchResourceDetails = async (resourceType, resource) => {
  const details = await (
    await fetch(`/api/${resourceType}/${resource}`)
  ).json();
  return details.data;
};

export const fetchResourceTypes = async () => {
  const res = await (await fetch("/api/resourceTypes")).json();
  return res.data;
};

export const fetchResourcesWithUserCount = async (resourceType) => {
  const details = await (
    await fetch(`/api/principals/${resourceType}`)
  ).json();
  return details.data;
};
