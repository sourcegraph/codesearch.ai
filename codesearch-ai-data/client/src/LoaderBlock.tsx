import React from "react";
import "./LoaderBlock.css";

export const LoaderBlock: React.FunctionComponent = () => {
  return (
    <div className="loader-block">
      <div className="loader-block-header">
        <div className="loader-block-body-line"></div>
      </div>
      <div className="loader-block-body">
        <div className="loader-block-body-line loader-block-body-line-short"></div>
        <div className="loader-block-body-line"></div>
        <div className="loader-block-body-line loader-block-body-line-short"></div>
        <div className="loader-block-body-line"></div>
        <div className="loader-block-body-line loader-block-body-line-short"></div>
      </div>
    </div>
  );
};
