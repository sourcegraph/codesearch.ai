import argparse

from tokenizers import ByteLevelBPETokenizer
from transformers import RobertaTokenizerFast

from codesearch_ai_ml.data_loader import stream_code_query_tuples
from codesearch_ai_ml.model_config import VOCAB_SIZE, MAX_LENGTH
from codesearch_ai_ml.util import flatten


TEXT_START_TOKEN = "<text>"
TEXT_STOP_TOKEN = "</text>"
CODE_START_TOKEN = "<code>"
CODE_STOP_TOKEN = "</code>"

SPECIAL_TOKENS = [
    "<s>",
    "<pad>",
    "</s>",
    "<unk>",
    "<mask>",
    TEXT_START_TOKEN,
    TEXT_STOP_TOKEN,
    CODE_START_TOKEN,
    CODE_STOP_TOKEN,
]


def get_code_tokenizer(model_path):
    return RobertaTokenizerFast.from_pretrained(
        model_path,
        add_prefix_space=True,
        cls_token=CODE_START_TOKEN,
        sep_token=CODE_STOP_TOKEN,
        max_len=MAX_LENGTH,
        local_files_only=True,
    )


def get_text_tokenizer(model_path):
    return RobertaTokenizerFast.from_pretrained(
        model_path,
        add_prefix_space=True,
        cls_token=TEXT_START_TOKEN,
        sep_token=TEXT_STOP_TOKEN,
        max_len=MAX_LENGTH,
        local_files_only=True,
    )


def tokenize(tokenizer, sample: str, padding=False, return_special_tokens_mask=True):
    return tokenizer(
        sample,
        padding=padding,
        add_special_tokens=True,
        return_special_tokens_mask=return_special_tokens_mask,
        truncation=True,
        return_tensors="pt",
    )


def train_tokenizer(code_query_pairs_file: str, output_dir: str):
    tokenizer = ByteLevelBPETokenizer(add_prefix_space=True)
    tokenizer.add_special_tokens(SPECIAL_TOKENS)
    tokenizer.add_tokens(
        [
            " " * 2,
            " " * 4,
            " " * 8,
            " " * 12,
            " " * 16,
            "\t",
            "\t" * 2,
            "\t" * 3,
            "\t" * 4,
        ]
    )
    tokenizer.train_from_iterator(
        (
            value
            for value in flatten(stream_code_query_tuples(code_query_pairs_file))
            if len(value) > 0
        ),
        vocab_size=VOCAB_SIZE,
        min_frequency=2,
        special_tokens=SPECIAL_TOKENS,
    )
    tokenizer.save_model(output_dir)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--code-query-pairs-file", dest="code_query_pairs_file")
    parser.add_argument("--output-dir", dest="output_dir")
    args = parser.parse_args()

    train_tokenizer(args.code_query_pairs_file, args.output_dir)
