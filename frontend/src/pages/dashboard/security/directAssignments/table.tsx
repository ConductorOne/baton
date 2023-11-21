import * as React from "react";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import { Link } from "react-router-dom";
import Arrow from "@mui/icons-material/ArrowCircleRightOutlined";

const rows = [
  {
    resourceName: "Credentials",
    resourceType: "vault",
    assignments: "10",
    link: "/valut/2343534",
  },
  {
    resourceName: "Cloud Keys",
    resourceType: "vault",
    assignments: "5",
    link: "/valut/342354",
  },
  {
    resourceName: "GitHub",
    resourceType: "repository",
    assignments: "200",
    link: "/respository/2343534",
  },
];

export const ResourcesTable = () => {
  return (
    <TableContainer>
      <Table sx={{ minWidth: 750 }}>
        <TableHead>
          <TableRow>
            <TableCell>RESOURCE NAME</TableCell>
            <TableCell align="right">TYPE</TableCell>
            <TableCell align="right">ASSIGNMENTS</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {rows.map((row) => (
            <TableRow
              key={row.resourceName}
              sx={{ "&:last-child td, &:last-child th": { border: 0 } }}
            >
              <TableCell component="th" scope="row">
                {row.resourceName}
              </TableCell>
              <TableCell align="right">{row.resourceType}</TableCell>
              <TableCell align="right">
                <Link to={row.link} style={{ textDecoration: "none", display: "inline-flex", justifyContent: "center", color: "inherit", alignSelf: "flex-end"}}>
                  {row.assignments}
                  <Arrow fontSize="small" sx={{marginLeft: "5px"}}/>
                </Link>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
