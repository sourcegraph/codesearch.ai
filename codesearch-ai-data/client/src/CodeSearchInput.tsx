import React, { useState } from "react";
import "./CodeSearchInput.css";

interface CodeSearchInputProps {
  query?: string;
  onSearch: (query: string) => void;
}

export const CodeSearchInput: React.FunctionComponent<CodeSearchInputProps> = ({
  query,
  onSearch,
}) => {
  const [value, setValue] = useState(query ?? "");

  return (
    <div className="code-search-input-root">
      <textarea
        placeholder="Enter a snippet of code..."
        className="code-search-input"
        value={value}
        onChange={(e) => setValue(e.target.value)}
      />
      <div>
        <button
          className="code-search-input-button"
          type="button"
          onClick={() => onSearch(value)}
        >
          Search
        </button>
      </div>
    </div>
  );
};
