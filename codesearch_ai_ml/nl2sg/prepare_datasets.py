import os
import argparse
from collections import defaultdict, OrderedDict
import random

import numpy as np
from scipy.special import softmax
from datasets import load_dataset, disable_caching

from codesearch_ai_ml.data_loader import stream_jsonl_file

N_RANDOMLY_SAMPLED_QUERIES = 2
MAX_CANDIDATE_TOKEN_SAMPLES = 10

columns_to_remove = [
    "id",
    "code",
    "tokens",
    "counts",
    "soQuestionId",
    "extractedFunctionId",
]


def tokens_to_and_query(tokens):
    if len(tokens) == 1:
        return tokens[0]
    return "({})".format(" AND ".join(tokens))


def generate_sourcegraph_queries(n_documents, token_counts_per_document, document):
    tokens, counts = document["tokens"], document["counts"]

    n_tokens = float(sum(counts))
    token_tf_idf_scores = [
        (
            token,
            # Term frequency
            (count / n_tokens)
            * (
                # Inverse document frequency
                np.log(
                    (
                        float(n_documents)
                        / (float(token_counts_per_document.get(token, 0.0)) + 1.0)
                    )
                )
                + 1.0
            ),
        )
        for token, count in zip(tokens, counts)
    ]

    tokens = [token for token, _ in token_tf_idf_scores]
    weights = list(softmax([tf_idf_score for _, tf_idf_score in token_tf_idf_scores]))
    tokens_ordered_by_tf_idf_score = sorted(
        token_tf_idf_scores, key=lambda x: x[1], reverse=True
    )

    n_sampled_tokens = min(len(tokens), MAX_CANDIDATE_TOKEN_SAMPLES)
    highest_scoring_tokens = frozenset(
        [token for token, _ in tokens_ordered_by_tf_idf_score[:n_sampled_tokens]]
    )

    candidate_queries = OrderedDict()
    candidate_queries[highest_scoring_tokens] = True

    for _ in range(N_RANDOMLY_SAMPLED_QUERIES):
        # Deduplicate tokens because we are sampling *with* replacement
        candidate_tokens = frozenset(
            random.choices(tokens, k=n_sampled_tokens, weights=weights)
        )
        candidate_queries[candidate_tokens] = True

    and_queries = []
    for tokens in candidate_queries.keys():
        and_queries.append(tokens_to_and_query(list(tokens)))

    return {"query": document["query"], "sourcegraph": " OR ".join(and_queries)}


def prepare_dataset(
    dataset, n_train_documents, token_counts_per_document, output_dir, num_proc
):
    processed_dataset = dataset.filter(
        lambda document: len(document["tokens"]) > 0 and len(document["query"]) > 0
    ).map(
        lambda document: generate_sourcegraph_queries(
            n_train_documents, token_counts_per_document, document
        ),
        num_proc=num_proc,
    )
    processed_dataset = processed_dataset.remove_columns(columns_to_remove)
    processed_dataset.save_to_disk(output_dir)


def prepare_datasets(
    train_file,
    test_file,
    output_dir,
    num_proc=16,
):
    token_counts_per_document = defaultdict(int)
    n_train_documents = 0

    for cqp in stream_jsonl_file(train_file):
        n_train_documents += 1

        for token in cqp["tokens"]:
            token_counts_per_document[token] += 1

    prepare_dataset(
        load_dataset("json", data_files=train_file, split="train"),
        n_train_documents,
        token_counts_per_document,
        os.path.join(output_dir, "train"),
        num_proc,
    )

    prepare_dataset(
        load_dataset("json", data_files=test_file, split="train"),
        n_train_documents,
        token_counts_per_document,
        os.path.join(output_dir, "test"),
        num_proc,
    )


if __name__ == "__main__":
    disable_caching()

    parser = argparse.ArgumentParser()
    parser.add_argument("--train-file", dest="train_file")
    parser.add_argument("--test-file", dest="test_file")
    parser.add_argument("--output-dir", dest="output_dir")
    parser.add_argument("--num-proc", dest="num_proc")
    args = parser.parse_args()

    print("Preparing datasets...")
    prepare_datasets(
        args.train_file,
        args.test_file,
        args.output_dir,
        num_proc=int(args.num_proc),
    )
