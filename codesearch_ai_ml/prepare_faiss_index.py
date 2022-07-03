import datetime
import os.path
import argparse
import multiprocessing as mp

import torch
import numpy as np
import faiss
from faiss.contrib.ondisk import merge_ondisk
from transformers import RobertaModel

from codesearch_ai_ml.data_loader import stream_code_query_tuples
from codesearch_ai_ml.encode import encode_code_samples
from codesearch_ai_ml.model_config import HIDDEN_SIZE
from codesearch_ai_ml.train_tokenizer import get_code_tokenizer
from codesearch_ai_ml.util import batches

CODE_BATCH_SIZE = 2**20
ENCODING_BATCH_SIZE = 2**10
FAISS_TRAINED_INDEX_NAME = "trained.index"


def train_faiss_index(encoded_code_batch: np.ndarray, output_dir: str):
    nlist = 4096 * 2  # Number of Voronoi cells
    m = 8
    bits = 8
    quantizer = faiss.IndexFlatIP(HIDDEN_SIZE)
    index = faiss.IndexIVFPQ(quantizer, HIDDEN_SIZE, nlist, m, bits)
    index.train(encoded_code_batch)
    faiss.write_index(index, os.path.join(output_dir, FAISS_TRAINED_INDEX_NAME))


def encode_and_train_faiss_index(
    code_batches_stream_fn,
    base_model_path: str,
    fine_tuned_model_path: str,
    output_dir: str,
):
    device = torch.device("cuda")
    code_tokenizer = get_code_tokenizer(base_model_path)
    encoder_model = RobertaModel.from_pretrained(fine_tuned_model_path)
    encoder_model.to(device)
    print(datetime.datetime.now(), "Encoding first batch...")
    encoded_first_code_batch = encode_code_batch(
        next(code_batches_stream_fn()), encoder_model, code_tokenizer, device
    )
    print(datetime.datetime.now(), "Training FAISS index...")
    train_faiss_index(encoded_first_code_batch, output_dir)

    # Free GPU memory
    del encoder_model
    torch.cuda.empty_cache()


def encode_code_batch(code_batch, encoder_model, code_tokenizer, device) -> np.ndarray:
    encoded_batches = []
    for encoding_batch in batches(code_batch, ENCODING_BATCH_SIZE):
        encoded_batches.append(
            encode_code_samples(encoder_model, code_tokenizer, encoding_batch, device)
        )
    return np.vstack(encoded_batches)


def mp_prepare_faiss_block(process_batch):
    idx, code_batch, base_model_path, fine_tuned_model_path, output_dir = process_batch

    cuda_idx = mp.current_process()._identity[0] - 1
    device = torch.device(f"cuda:{cuda_idx}")

    print(datetime.datetime.now(), "Started processing batch", idx, cuda_idx)

    code_tokenizer = get_code_tokenizer(base_model_path)
    encoder_model = RobertaModel.from_pretrained(fine_tuned_model_path)
    encoder_model.to(device)

    start_idx = CODE_BATCH_SIZE * idx
    index = faiss.read_index(os.path.join(output_dir, FAISS_TRAINED_INDEX_NAME))
    vecs = encode_code_batch(code_batch, encoder_model, code_tokenizer, device)
    index.add_with_ids(vecs, np.arange(start_idx, start_idx + vecs.shape[0]))
    faiss.write_index(index, os.path.join(output_dir, f"block_{idx}.index"))

    print(datetime.datetime.now(), "Done processing batch", idx, cuda_idx)


def prepare_faiss_index(
    base_model_path: str,
    fine_tuned_model_path: str,
    code_query_pairs_file: str,
    output_dir: str,
    n_gpus=4,
):
    code_batches_stream_fn = lambda: batches(
        (code for code, _ in stream_code_query_tuples(code_query_pairs_file)),
        CODE_BATCH_SIZE,
    )

    encode_and_train_faiss_index(
        code_batches_stream_fn, base_model_path, fine_tuned_model_path, output_dir
    )

    code_batches = [
        (idx, batch, base_model_path, fine_tuned_model_path, output_dir)
        for idx, batch in enumerate(code_batches_stream_fn())
    ]

    with mp.Pool(n_gpus) as p:
        p.map(mp_prepare_faiss_block, code_batches)

    index = faiss.read_index(os.path.join(output_dir, FAISS_TRAINED_INDEX_NAME))
    block_fnames = [
        os.path.join(output_dir, f"block_{i}.index") for i in range(len(code_batches))
    ]
    merge_ondisk(index, block_fnames, os.path.join(output_dir, "merged_index.ivfdata"))
    faiss.write_index(index, os.path.join(output_dir, "populated.index"))


if __name__ == "__main__":
    torch.multiprocessing.set_start_method("spawn", force=True)
    mp.set_start_method("spawn", force=True)

    parser = argparse.ArgumentParser()
    parser.add_argument("--base-model", dest="base_model")
    parser.add_argument("--fine-tuned-model", dest="fine_tuned_model")
    parser.add_argument("--code-query-pairs-file", dest="code_query_pairs_file")
    parser.add_argument("--output-dir", dest="output_dir")
    parser.add_argument("--n-gpus", dest="n_gpus")
    args = parser.parse_args()

    prepare_faiss_index(
        args.base_model,
        args.fine_tuned_model,
        args.code_query_pairs_file,
        args.output_dir,
        n_gpus=int(args.n_gpus),
    )
