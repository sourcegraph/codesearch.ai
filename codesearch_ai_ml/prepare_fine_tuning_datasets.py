import argparse

from datasets import load_from_disk


def prepare_fine_tuning_dataset(dataset_path: str, output_path: str, num_proc=16):
    dataset = load_from_disk(dataset_path)
    dataset.cleanup_cache_files()
    pretraining_dataset = dataset.filter(
        lambda sample: sample["query_tokenized"] is not None
    ).map(
        lambda sample: {
            "code_input_ids": sample["code_tokenized"]["input_ids"],
            "code_attention_mask": sample["code_tokenized"]["attention_mask"],
            "query_input_ids": sample["query_tokenized"]["input_ids"],
            "query_attention_mask": sample["query_tokenized"]["attention_mask"],
        },
        num_proc=num_proc,
    )
    # TODO: Temporary code, query removal
    pretraining_dataset = pretraining_dataset.remove_columns(
        ["code_tokenized", "query_tokenized"]
    )
    pretraining_dataset.save_to_disk(output_path)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--dataset", dest="dataset")
    parser.add_argument("--output-dir", dest="output_dir")
    parser.add_argument("--num-proc", dest="num_proc")
    args = parser.parse_args()

    print("Preparing fine-tuning dataset...")
    prepare_fine_tuning_dataset(
        args.dataset, args.output_dir, num_proc=int(args.num_proc)
    )
