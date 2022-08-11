import { useEffect, useState } from "react";

const BASE_API_URL = process.env.REACT_APP_BASE_API_URL || "";

export function useSearchResults<T>(
  source: "functions" | "so",
  by: "text" | "code",
  query: string
): T[] | Error | "loading" | null {
  const [results, setResults] = useState<T[] | Error | "loading" | null>(null);

  useEffect(() => {
    setResults("loading");

    fetch(
      `${BASE_API_URL}/api/search/${source}/by-${by}?query=${encodeURIComponent(
        query
      )}`
    )
      .then((response) => response.json())
      .then((response) => setResults(response as T[]))
      .catch((error) => setResults(new Error(error)));
  }, [source, by, query, setResults]);

  return results;
}
