# codesearch.ai

[codesearch.ai](https://codesearch.ai) is a semantic code search engine. It allows searching GitHub functions and StackOverflow answers using natural language queries. It uses HuggingFace Transformers under the hood, and the training procedure is inspired by a paper called [Text and Code Embeddings by Contrastive Pre-Training](https://arxiv.org/pdf/2201.10005.pdf) from OpenAI. The [CodeSearchNet project](https://github.com/github/CodeSearchNet) served as a basis for data collection and cleaning.

The project is split into two sub-projects: data collection and model training. The `codesearch-ai-data` folder corresponds to the data collection part written in Go. And the `codesearch_ai_ml` folder corresponds to the model training part written in Python.

## Requirements

- Go >= 1.18
- Python >= 3.7
- CUDA (for GPU model training)
- Postgres

## Code walkthrough

We prepared a detailed code walkthrough in the form of a [Sourcegraph Notebook](https://sourcegraph.com/notebooks/Tm90ZWJvb2s6MTM0Mw==).
