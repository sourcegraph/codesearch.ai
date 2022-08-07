import React, { useMemo } from "react";
import "simplebar-react/dist/simplebar.min.css";
import SimpleBar from "simplebar-react";

import githubMark from "./svgs/github-mark.svg";
import "./CodeSnippet.css";
import "./highlight.css";

interface CodeSnippetProps {
  repositoryName: string;
  filePath: string;
  highlightedHTML: string;
  url: string;
}

export const CodeSnippet: React.FunctionComponent<CodeSnippetProps> = ({
  repositoryName,
  filePath,
  highlightedHTML,
  url,
}) => {
  const fileName = useMemo(() => {
    const filePathSplit = filePath.split("/");
    return filePathSplit[filePathSplit.length - 1];
  }, [filePath]);

  const repoistoryNameStripped = useMemo(() => {
    if (repositoryName.startsWith("github.com/")) {
      return repositoryName.slice("github.com/".length);
    }
    return repositoryName;
  }, [repositoryName]);

  return (
    <div className="code-snippet">
      <div className="code-snippet-header">
        <img src={githubMark} alt="GitHub Mark" width="16" height="17" />
        <a href={url}>
          {repoistoryNameStripped} &middot; <strong>{fileName}</strong>
        </a>
      </div>
      <SimpleBar style={{ maxHeight: 500 }}>
        <div className="code-snippet-highlighted-code">
          <pre>
            <div dangerouslySetInnerHTML={{ __html: highlightedHTML }}></div>
          </pre>
        </div>
      </SimpleBar>
    </div>
  );
};
