import React from "react";
import { CodeSnippet } from "./CodeSnippet";
import { useSearchResults } from "./useSearchResults";
import { HighlightedFunction, SOQuestion } from "./types";
import { SOQuestionComponent } from "./SOQuestionComponent";
import { LoaderBlock } from "./LoaderBlock";

import "simplebar-react/dist/simplebar.min.css";
import "./SearchResults.css";

export const isErrorLike = (value: unknown): value is Error =>
  typeof value === "object" &&
  value !== null &&
  ("stack" in value || "message" in value) &&
  !("__typename" in value);

export const SearchResults: React.FunctionComponent<{
  query: string;
  searchBy: "text" | "code";
}> = ({ query, searchBy }) => {
  const functionsSearchResults = useSearchResults<HighlightedFunction>(
    "functions",
    searchBy,
    query
  );

  const soSearchResults = useSearchResults<SOQuestion>("so", searchBy, query);

  const loadingBlocks = (
    <>
      <LoaderBlock />
      <LoaderBlock />
      <LoaderBlock />
    </>
  );

  return (
    <div className="search-results">
      <div className="search-results-column">
        {functionsSearchResults === "loading" && loadingBlocks}
        {functionsSearchResults &&
          functionsSearchResults !== "loading" &&
          !isErrorLike(functionsSearchResults) &&
          functionsSearchResults.map((result) => (
            <CodeSnippet key={`function-${result.id}`} {...result} />
          ))}
      </div>
      <div className="search-results-column">
        {soSearchResults === "loading" && loadingBlocks}
        {soSearchResults &&
          soSearchResults !== "loading" &&
          !isErrorLike(soSearchResults) &&
          soSearchResults.map((result) => (
            <SOQuestionComponent key={`so-question-${result.id}`} {...result} />
          ))}
      </div>
    </div>
  );
};
