import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import "./index.css";
import { HomePage } from "./HomePage";
import { SearchResultsPage } from "./SearchResultsPage";

const root = ReactDOM.createRoot(
  document.getElementById("root") as HTMLElement
);

root.render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route
          path="/search/by-text"
          element={<SearchResultsPage searchBy="text" />}
        />
        <Route
          path="/search/by-code"
          element={<SearchResultsPage searchBy="code" />}
        />
        <Route path="*" element={<div>Page not found.</div>} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
);
