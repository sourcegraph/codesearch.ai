import React from "react";
import "./QueryExampleChip.css";

export const QueryExampleChip: React.FunctionComponent<{
  url: string;
  text: string;
}> = ({ url, text }) => {
  return (
    <a href={url} className="query-example-chip">
      {text}
    </a>
  );
};
