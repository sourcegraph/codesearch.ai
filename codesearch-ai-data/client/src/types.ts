export interface HighlightedFunction {
  id: number;
  repositoryName: string;
  commitID: string;
  filePath: string;
  startLine: number;
  endLine: number;
  highlightedHTML: string;
  url: string;
}

export interface SOQuestion {
  id: number;
  title: string;
  tags: string;
  creationDate: string;
  score: number;
  answers: SOAnswer[];
  url: string;
}

interface SOAnswer {
  id: string;
  body: string;
  score: number;
  creation_date: string;
}
