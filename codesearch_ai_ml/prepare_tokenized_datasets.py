import argparse
from datasets import load_dataset

from codesearch_ai_ml.util import extract_first_tensor
from codesearch_ai_ml.train_tokenizer import tokenize, get_code_tokenizer, get_text_tokenizer


columns_to_remove = [
    "id",
    "code",
    "query",
    "soQuestionId",
    "extractedFunctionId",
]


def tokenize_code_and_query(code_tokenizer, text_tokenizer, sample):
    sample["code_tokenized"] = extract_first_tensor(
        tokenize(code_tokenizer, sample["code"])
    )
    sample["code_length"] = len(sample["code_tokenized"]["input_ids"])

    if len(sample["query"]) > 0:
        sample["query_tokenized"] = extract_first_tensor(
            tokenize(text_tokenizer, sample["query"])
        )
        sample["query_length"] = len(sample["query_tokenized"]["input_ids"])
    else:
        sample["query_tokenized"] = None
        sample["query_length"] = 0
    return sample


def prepare_tokenized_dataset(
    model_path,
    code_query_pairs_file,
    output_dir,
    num_proc=16,
):
    code_tokenizer = get_code_tokenizer(model_path)
    text_tokenizer = get_text_tokenizer(model_path)

    dataset = load_dataset("json", data_files=code_query_pairs_file, split="train")
    dataset = dataset.map(
        lambda sample: tokenize_code_and_query(code_tokenizer, text_tokenizer, sample),
        num_proc=num_proc,
    )
    dataset = dataset.remove_columns(columns_to_remove)
    dataset.save_to_disk(output_dir)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--model-path", dest="model_path")
    parser.add_argument("--code-query-pairs-file", dest="code_query_pairs_file")
    parser.add_argument("--output-dir", dest="output_dir")
    parser.add_argument("--num-proc", dest="num_proc")
    args = parser.parse_args()

    print("Preparing tokenized dataset...")
    prepare_tokenized_dataset(
        args.model_path,
        args.code_query_pairs_file,
        args.output_dir,
        num_proc=int(args.num_proc),
    )
