import React from "react";
import { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import "./Home.css";
import { Logo } from "./Logo";
import { QueryExampleChip } from "./QueryExampleChip";
import { TextSearchInput } from "./TextSearchInput";

const textSearchExamples = [
  "how to determine a string is a valid word",
  "how to get a database table name",
  "convert a date string into yyyymmdd",
  "how to make the checkbox checked",
  "how to read a .gz compressed file?",
];

export const Home: React.FunctionComponent<{}> = () => {
  const navigate = useNavigate();

  const onSearch = useCallback(
    (query: string) => {
      navigate(`/search/by-text?query=${encodeURIComponent(query)}`);
    },
    [navigate]
  );

  return (
    <div className="home-container">
      <div className="home">
        <div className="home-logo">
          <Logo />
        </div>
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
      </div>
    </div>
  );
};
