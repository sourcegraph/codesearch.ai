import json
import argparse

from codesearch_ai_ml.data_loader import stream_jsonl_file


def prepare_index_id_mapping(
    code_query_pairs_file: str, id_column: str, output_file: str
):
    mapping = [cqp[id_column] for cqp in stream_jsonl_file(code_query_pairs_file)]

    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(mapping, f)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--code-query-pairs-file", dest="code_query_pairs_file")
    parser.add_argument("--id-column", dest="id_column")
    parser.add_argument("--output-file", dest="output_file")
    args = parser.parse_args()

    print("Preparing mapping...")
    prepare_index_id_mapping(
        args.code_query_pairs_file,
        args.id_column,
        args.output_file,
    )
