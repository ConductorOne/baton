import React, { useEffect } from "react";
import {
  BaseEdge,
  EdgeLabelRenderer,
  EdgeProps,
  getBezierPath,
} from "reactflow";
import { FormControl, Select } from "@mui/material";
import MenuItem from "@mui/material/MenuItem";
import {
  EntitlementNumberLabel,
  SelectedEntitlementWrapper,
} from "../styles/styles";
import { colors } from "../../../style/colors"

const EntitlementsMenu = ({ entitlements, openEntitlementsDetails, opacity }) => {
  const [entitlement, setEntitlement] = React.useState<string>(
    entitlements[0].slug
  );
  const handleChange = (event) => {
    setEntitlement(event.target.value);
  };

  useEffect(() => {
    // this prevents opening the menu on reload
    if (entitlement !== entitlements[0].slug) {
      const selectedEntitlement = entitlements.find(
        (obj) => obj.slug === entitlement
      );
      if (selectedEntitlement) {
        openEntitlementsDetails(selectedEntitlement);
      }
    }
  }, [entitlement]);

  const multipleEntitlements = entitlements.length > 1;
  return (
    <FormControl>
      <Select
        variant="standard"
        labelId={entitlement}
        id="entitlements-select"
        value={entitlement}
        onChange={handleChange}
        MenuProps={multipleEntitlements ? {} : { open: false }}
        disableUnderline
        renderValue={() => (
          <SelectedEntitlementWrapper sx={opacity}>
            {multipleEntitlements && (
              <EntitlementNumberLabel>
                {entitlements.length}
              </EntitlementNumberLabel>
            )}
            {entitlement}
          </SelectedEntitlementWrapper>
        )}
      >
        {multipleEntitlements &&
          entitlements.map((ent) => (
            <MenuItem key={ent.id} value={ent.slug}>
              {ent.slug}
            </MenuItem>
          ))}
      </Select>
    </FormControl>
  );
};

export const CustomEdge = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style,
  markerEnd,
  data,
}: EdgeProps) => {
  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  const baseStyle = {
    ...data.style,
    stroke: colors.batonGreen600,
  };

  return (
    <>
      <BaseEdge path={edgePath} markerEnd={markerEnd} style={baseStyle} id={id} />
      <EdgeLabelRenderer>
        <div
          style={{
            position: "absolute",
            transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
            pointerEvents: "all",
          }}
        >
          {data.entitlements && (
            <EntitlementsMenu
              entitlements={data.entitlements}
              openEntitlementsDetails={data.openEntitlementsDetails}
              opacity={data.style}
            />
          )}
        </div>
      </EdgeLabelRenderer>
    </>
  );
};
