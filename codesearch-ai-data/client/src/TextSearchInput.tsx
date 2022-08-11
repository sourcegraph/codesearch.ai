import React, { useState } from "react";
import searchIcon from "./svgs/search-icon.svg";
import "./TextSearchInput.css";

interface TextSearchInputProps {
  query?: string;
  onSearch: (query: string) => void;
}

export const TextSearchInput: React.FunctionComponent<TextSearchInputProps> = ({
  query,
  onSearch,
}) => {
  const [value, setValue] = useState(query ?? "");

  return (
    <div className="text-search-input-root">
      <img
        src={searchIcon}
        width="16"
        height="16"
        alt="Search Icon"
        className="text-search-input-icon"
      />
      <input
        type="text"
        placeholder="Enter a search query..."
        className="text-search-input"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            onSearch(value);
          }
        }}
        required
      />
    </div>
  );
};
