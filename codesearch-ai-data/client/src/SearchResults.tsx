import React, { useEffect, useMemo, useState } from "react";
import { useLocation } from "react-router-dom";

import githubMark from "./github-mark.png";
import soLogo from "./stack-overflow.png";

import "./highlight.css";
import "./SearchResults.css";

function getSearchResultsURL(
  query: string | null,
  code: string | null,
  language: string | null,
  set: "extracted-functions" | "so"
): string | null {
  if (query) {
    return `/api/search/${set}?query=${encodeURIComponent(query)}`;
  } else if (code && language) {
    return `/api/search/${set}?code=${encodeURIComponent(
      code
    )}&language=${language}`;
  }
  return null;
}

interface HighlightedExtractedFunction {
  id: number;
  repositoryName: string;
  commitID: string;
  filePath: string;
  startLine: number;
  endLine: number;
  highlightedHTML: string;
  url: string;
}

interface SOQuestionWithAnswers {
  id: number;
  title: string;
  tags: string;
  creationDate: string;
  score: number;
  answers: { id: string; body: string; score: string; creationDate: string }[];
  url: string;
}

async function getSearchResults<T>(url: string | null): Promise<T | null> {
  if (!url) {
    return null;
  }
  return fetch(url)
    .then((response) => response.json())
    .catch(() => {
      return null;
    });
}

export const SearchResults: React.FunctionComponent<{}> = () => {
  const location = useLocation();

  const [extractedFunctionsSearchResults, setExtractedFunctionsSearchResults] =
    useState<HighlightedExtractedFunction[] | "loading" | null>("loading");
  const [soSearchResults, setSOSearchResults] = useState<
    SOQuestionWithAnswers[] | "loading" | null
  >("loading");

  const queryParams = useMemo(
    () => new URLSearchParams(location.search),
    [location.search]
  );

  const query = useMemo(() => queryParams.get("query"), [queryParams]);
  const code = useMemo(() => queryParams.get("code"), [queryParams]);
  const language = useMemo(() => queryParams.get("language"), [queryParams]);

  const extractedFunctionsSearchResultsURL = useMemo(
    () => getSearchResultsURL(query, code, language, "extracted-functions"),
    [query, code, language]
  );
  const soSearchResultsURL = useMemo(
    () => getSearchResultsURL(query, code, language, "so"),
    [query, code, language]
  );

  useEffect(() => {
    getSearchResults<HighlightedExtractedFunction[]>(
      extractedFunctionsSearchResultsURL
    ).then((results) => setExtractedFunctionsSearchResults(results));

    getSearchResults<SOQuestionWithAnswers[]>(soSearchResultsURL).then(
      (results) => setSOSearchResults(results)
    );
  }, [
    extractedFunctionsSearchResultsURL,
    soSearchResultsURL,
    setExtractedFunctionsSearchResults,
    setSOSearchResults,
  ]);

  return (
    <div className="content">
      <div className="results">
        <div className="functions">
          <div className="functions__title">
            <img
              src={githubMark}
              width="16px"
              height="16px"
              alt="Github mark"
            />
            GitHub functions
          </div>
          {extractedFunctionsSearchResults &&
            extractedFunctionsSearchResults !== "loading" &&
            extractedFunctionsSearchResults.map((fn) => (
              <HighlightedExtractedFunctionComponent
                key={`${fn.repositoryName}${fn.filePath}${fn.startLine}${fn.endLine}`}
                {...fn}
              />
            ))}
        </div>
        <div className="questions">
          <div className="functions__title">
            <img
              src={soLogo}
              width="16px"
              height="16px"
              alt="StackOverflow logo"
            />
            StackOverflow (Experimental)
          </div>
          {soSearchResults &&
            soSearchResults !== "loading" &&
            soSearchResults.map((q) => (
              <SOQuestionWithAnswersComponent key={q.id} {...q} />
            ))}
        </div>
      </div>
    </div>
  );
};

const HighlightedExtractedFunctionComponent: React.FunctionComponent<
  HighlightedExtractedFunction
> = ({ repositoryName, filePath, highlightedHTML, url }) => (
  <div className="function">
    <a className="function__title" href={url}>
      {repositoryName}/{filePath}
    </a>
    <div className="function__code__wrapper">
      <div
        className="function__code"
        dangerouslySetInnerHTML={{ __html: highlightedHTML }}
      ></div>
      <button type="button" className="expand">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          className="feather feather-chevron-down"
        >
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </button>
    </div>
  </div>
);

const SOQuestionWithAnswersComponent: React.FunctionComponent<
  SOQuestionWithAnswers
> = ({ url, title, score, creationDate, answers }) => (
  <div className="question">
    <a className="question__title" href="{{ .URL }}">
      {title}
    </a>
    <div className="question__subtitle">
      Asked on {creationDate} &middot;
      <strong>{score}</strong> votes
    </div>
    <div className="question__answers__wrapper">
      <div className="question__answers__title">
        {answers.length}
        {answers.length === 1 ? "answer" : "answers"}
      </div>
      <div className="question__answers">
        {answers
          .filter((a) => !!a)
          .map((a) => (
            <div className="question__answer">
              <div className="question__answer__meta">
                Answered on {a.creationDate} &middot;
                <strong>{a.score}</strong> votes
              </div>
              <div
                className="question__answer__body"
                dangerouslySetInnerHTML={{ __html: a.body }}
              ></div>
            </div>
          ))}
      </div>
      <button type="button" className="expand expand-question">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          className="feather feather-chevron-down"
        >
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </button>
    </div>
  </div>
);
