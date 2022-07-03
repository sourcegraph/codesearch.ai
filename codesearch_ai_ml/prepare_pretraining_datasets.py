import argparse

from datasets import load_from_disk, concatenate_datasets


columns_to_remove = ["code_tokenized", "code_length", "query_tokenized", "query_length"]


def map_tokenized_sample(sample, prefix):
    return {**sample[f"{prefix}_tokenized"], "length": sample[f"{prefix}_length"]}


def get_code_dataset(dataset, num_proc=16):
    code_dataset = dataset.map(
        lambda sample: map_tokenized_sample(sample, "code"), num_proc=num_proc
    )
    return code_dataset


def get_queries_dataset(dataset, num_proc=16):
    queries_dataset = dataset.filter(
        lambda sample: sample["query_tokenized"] is not None,
    ).map(
        lambda sample: map_tokenized_sample(sample, "query"),
        num_proc=num_proc,
    )
    return queries_dataset


def get_pretraining_dataset(dataset, num_proc=16):
    return concatenate_datasets(
        [
            get_code_dataset(dataset, num_proc=num_proc),
            get_queries_dataset(dataset, num_proc=num_proc),
        ]
    )


def prepare_pretraining_dataset(dataset_path: str, output_path: str, num_proc=16):
    dataset = load_from_disk(dataset_path)
    dataset.cleanup_cache_files()
    pretraining_dataset = get_pretraining_dataset(dataset, num_proc=num_proc)
    pretraining_dataset = pretraining_dataset.remove_columns(columns_to_remove)
    pretraining_dataset.save_to_disk(output_path)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--dataset", dest="dataset")
    parser.add_argument("--output-dir", dest="output_dir")
    parser.add_argument("--num-proc", dest="num_proc")
    args = parser.parse_args()

    print("Preparing pretraining dataset...")
    prepare_pretraining_dataset(
        args.dataset, args.output_dir, num_proc=int(args.num_proc)
    )
