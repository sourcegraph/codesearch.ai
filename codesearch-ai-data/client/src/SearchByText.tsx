import React, { useCallback, useMemo } from "react";
import SimpleBar from "simplebar-react";

import { CodeSnippet } from "./CodeSnippet";
import { Logo } from "./Logo";
import { TextSearchInput } from "./TextSearchInput";

import "simplebar-react/dist/simplebar.min.css";
import "./SearchByText.css";
import { useSearchResults } from "./useSearchResults";
import { useLocation, useNavigate } from "react-router-dom";
import { HighlightedFunction, SOQuestion } from "./types";
import { SOQuestionComponent } from "./SOQuestionComponent";
import { LoaderBlock } from "./LoaderBlock";

export const isErrorLike = (value: unknown): value is Error =>
  typeof value === "object" &&
  value !== null &&
  ("stack" in value || "message" in value) &&
  !("__typename" in value);

export const SearchByText: React.FunctionComponent = () => {
  const location = useLocation();
  const navigate = useNavigate();

  const onSearch = useCallback(
    (query: string) => {
      navigate(`/search/by-text?query=${encodeURIComponent(query)}`);
    },
    [navigate]
  );

  const query = useMemo(() => {
    const searchParams = new URLSearchParams(location.search);
    return searchParams.get("query") ?? "";
  }, [location.search]);

  const functionsSearchResults = useSearchResults<HighlightedFunction>(
    "functions",
    "text",
    query
  );

  const soSearchResults = useSearchResults<SOQuestion>("so", "text", query);

  const loadingBlocks = (
    <>
      <LoaderBlock />
      <LoaderBlock />
      <LoaderBlock />
    </>
  );

  return (
    <SimpleBar style={{ height: "100vh" }}>
      <div className="search-by-text">
        <div className="search-by-text-header">
          <a href="/">
            <Logo />
          </a>
          <TextSearchInput query={query} onSearch={onSearch} />
        </div>
        <div className="search-by-text-results">
          <div className="search-by-text-results-column">
            {functionsSearchResults === "loading" && loadingBlocks}
            {functionsSearchResults &&
              functionsSearchResults !== "loading" &&
              !isErrorLike(functionsSearchResults) &&
              functionsSearchResults.map((result) => (
                <CodeSnippet key={`function-${result.id}`} {...result} />
              ))}
          </div>
          <div className="search-by-text-results-column">
            {soSearchResults === "loading" && loadingBlocks}
            {soSearchResults &&
              soSearchResults !== "loading" &&
              !isErrorLike(soSearchResults) &&
              soSearchResults.map((result) => (
                <SOQuestionComponent
                  key={`so-question-${result.id}`}
                  {...result}
                />
              ))}
          </div>
        </div>
      </div>
    </SimpleBar>
  );
};
