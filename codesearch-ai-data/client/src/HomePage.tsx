import React, { useState } from "react";
import { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { CodeSearchInput } from "./CodeSearchInput";
import { Logo } from "./Logo";
import { QueryExampleChip } from "./QueryExampleChip";
import { TextSearchInput } from "./TextSearchInput";

import "./HomePage.css";

const textSearchExamples = [
  "how to determine a string is a valid word",
  "how to get a database table name",
  "convert a date string into yyyymmdd",
  "how to make the checkbox checked",
  "how to read a .gz compressed file?",
];

export const HomePage: React.FunctionComponent<{}> = () => {
  const navigate = useNavigate();

  const [searchBy, setSearchBy] = useState<"text" | "code">("text");

  const onSearch = useCallback(
    (query: string) => {
      navigate(`/search/by-${searchBy}?query=${encodeURIComponent(query)}`);
    },
    [searchBy, navigate]
  );

  return (
    <div className="home-container">
      <div className="home">
        <div className="home-logo">
          <Logo />
        </div>
        <div className="home-search-by">
          <button
            className={`home-search-by-option ${
              searchBy === "text" && "home-search-by-option-selected"
            }`}
            type="button"
            onClick={() => setSearchBy("text")}
          >
            Search by text
          </button>
          <button
            className={`home-search-by-option ${
              searchBy === "code" && "home-search-by-option-selected"
            }`}
            type="button"
            onClick={() => setSearchBy("code")}
          >
            Search by code
          </button>
        </div>
        {searchBy === "text" && (
          <>
            <div className="home-text-search-input">
              <TextSearchInput onSearch={onSearch} />
            </div>
            <div className="text-search-examples">
              {textSearchExamples.map((example) => (
                <QueryExampleChip
                  key={example}
                  url={`/search/by-text?query=${encodeURIComponent(example)}`}
                  text={example}
                />
              ))}
            </div>
          </>
        )}
        {searchBy === "code" && (
          <>
            <div className="home-code-search-input">
              <CodeSearchInput onSearch={onSearch} />
            </div>
          </>
        )}
      </div>
    </div>
  );
};
