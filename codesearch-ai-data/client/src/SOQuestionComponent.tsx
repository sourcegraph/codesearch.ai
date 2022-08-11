import React from "react";
import "simplebar-react/dist/simplebar.min.css";
import SimpleBar from "simplebar-react";
import { SOQuestion } from "./types";

import soIcon from "./svgs/so-icon.svg";
import "./SOQuestionComponent.css";

const scoreToString = (score: number): string =>
  score === 1 ? "1 point" : `${score} points`;

export const SOQuestionComponent: React.FunctionComponent<SOQuestion> = ({
  title,
  creationDate,
  score,
  answers,
  url,
}) => {
  return (
    <div className="so-question">
      <div className="so-question-header">
        <img src={soIcon} alt="StackOverflow Icon" width="16" height="17" />
        <div className="so-question-header-title-meta-wrapper">
          <a href={url} className="so-question-header-title">
            <strong>{title}</strong>
          </a>
          <div className="so-question-header-meta">
            Asked on {creationDate} &middot; {scoreToString(score)}
          </div>
        </div>
      </div>
      <SimpleBar style={{ maxHeight: 500 }}>
        <div className="so-question-answers">
          {answers.map((answer) => (
            <div className="so-question-answer" key={`answer-${answer.id}`}>
              <div className="so-question-answer-title">
                Answered on {answer.creation_date} &middot;{" "}
                {scoreToString(answer.score)}
              </div>
              <div dangerouslySetInnerHTML={{ __html: answer.body }} />
            </div>
          ))}
        </div>
      </SimpleBar>
    </div>
  );
};
