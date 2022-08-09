import React, { useCallback, useEffect, useMemo } from "react";
import SimpleBar from "simplebar-react";
import { Logo } from "./Logo";
import { TextSearchInput } from "./TextSearchInput";
import { useLocation, useNavigate } from "react-router-dom";
import { SearchResults } from "./SearchResults";

import "./SearchResultsPage.css";
import { CodeSearchInput } from "./CodeSearchInput";

export const SearchResultsPage: React.FunctionComponent<{
  searchBy: "text" | "code";
}> = ({ searchBy }) => {
  const location = useLocation();
  const navigate = useNavigate();

  const onSearch = useCallback(
    (query: string) => {
      navigate(`/search/by-${searchBy}?query=${encodeURIComponent(query)}`);
    },
    [searchBy, navigate]
  );

  const query = useMemo(() => {
    const searchParams = new URLSearchParams(location.search);
    return searchParams.get("query") ?? "";
  }, [location.search]);

  useEffect(() => {
    const querySummary = query.length > 64 ? `${query.slice(0, 64)}...` : query;
    document.title = `codesearch.ai | ${querySummary}`;
  }, [query]);

  return (
    <SimpleBar style={{ height: "100vh" }}>
      <div className="search-results-page">
        <div className="search-results-page-header">
          <a href="/">
            <Logo />
          </a>
          {searchBy === "text" ? (
            <TextSearchInput query={query} onSearch={onSearch} />
          ) : (
            <CodeSearchInput query={query} onSearch={onSearch} />
          )}
        </div>
        <SearchResults query={query} searchBy={searchBy} />
      </div>
    </SimpleBar>
  );
};
