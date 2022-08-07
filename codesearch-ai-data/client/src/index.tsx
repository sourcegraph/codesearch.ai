import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import "./index.css";
import { Home } from "./Home";
import { SearchByText } from "./SearchByText";

const root = ReactDOM.createRoot(
  document.getElementById("root") as HTMLElement
);

root.render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/search/by-text" element={<SearchByText />} />
        <Route path="*" element={<div>Page not found.</div>} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
);
