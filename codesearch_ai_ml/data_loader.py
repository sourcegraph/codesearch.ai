import json
from typing import Tuple


def stream_jsonl_file(file_path: str):
    with open(file_path, encoding="utf-8") as f:
        for line in f:
            yield json.loads(line)


def stream_code_query_tuples(file_path: str) -> Tuple[str, str]:
    return ((cqp["code"], cqp["query"]) for cqp in stream_jsonl_file(file_path))
