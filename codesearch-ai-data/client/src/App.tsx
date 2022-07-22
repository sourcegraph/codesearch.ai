import classNames from "classnames";
import React, { useCallback, useState } from "react";
import { useNavigate } from "react-router-dom";
import "./App.css";

export const App: React.FunctionComponent<{}> = () => {
  const [searchBy, setSearchBy] = useState<"text" | "code">("text");
  const [language, setLanguage] = useState("python");
  const [query, setQuery] = useState("");
  const [code, setCode] = useState("");

  const navigate = useNavigate();

  const search = useCallback(() => {
    console.log(searchBy, language, query, code);
    if (searchBy === "text") {
      navigate(`/search?query=${encodeURIComponent(query)}`);
    } else {
      navigate(`/search?code=${encodeURIComponent(code)}&language=${language}`);
    }
  }, [searchBy, language, query, code, navigate]);

  return (
    <div className="app">
      <h1>codesearch.ai</h1>
      <div className="tab-buttons">
        <button
          className={classNames(
            "tab-button",
            searchBy === "text" && "tab-button--active"
          )}
          type="button"
          onClick={() => setSearchBy("text")}
        >
          Search using natural language
        </button>
        <button
          className={classNames(
            "tab-button",
            searchBy === "code" && "tab-button--active"
          )}
          type="button"
          onClick={() => setSearchBy("code")}
        >
          Search using code
        </button>
      </div>
      {searchBy === "text" && (
        <div className="text-input">
          <input
            type="text"
            placeholder="Enter query"
            className="text-input__input"
            onChange={(e) => setQuery(e.target.value)}
          />
          <button type="button" className="search-button" onClick={search}>
            Search
          </button>
        </div>
      )}
      {searchBy === "code" && (
        <div className="code-input">
          <textarea
            placeholder="Enter code"
            className="code-input__textarea"
            onChange={(e) => setCode(e.target.value)}
          ></textarea>
          <select
            className="code-input__select-language"
            value={language}
            onChange={(e) => setLanguage(e.target.value)}
          >
            <option value="python">Python</option>
            <option value="go">Go</option>
            <option value="ruby">Ruby</option>
            <option value="java">Java</option>
            <option value="javascript">Javascript</option>
            <option value="php">PHP</option>
          </select>
          <button type="button" className="search-button" onClick={search}>
            Search
          </button>
        </div>
      )}
    </div>
  );
};
