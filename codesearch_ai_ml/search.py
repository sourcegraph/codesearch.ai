import os
import argparse

import torch
import faiss
from transformers import RobertaModel

from codesearch_ai_ml.encode import encode_code_samples
from codesearch_ai_ml.train_tokenizer import get_text_tokenizer

device = torch.device("cuda")


def find_similar_vector_ids(faiss_index, query_vectors, n_results=5):
    return faiss_index.search(query_vectors, n_results)


def search_code_documents(query, encoder_model, faiss_index, tokenizer, n_results=20):
    encoded_query = encode_code_samples(encoder_model, tokenizer, [query], device)
    _, indices = find_similar_vector_ids(
        faiss_index, encoded_query, n_results=n_results
    )
    return indices[0]


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-model", dest="base_model")
    parser.add_argument("--fine-tuned-model", dest="fine_tuned_model")
    parser.add_argument("--faiss-dir", dest="faiss_dir")
    parser.add_argument("--q", dest="query")
    args = parser.parse_args()

    query_tokenizer = get_text_tokenizer(args.base_model)
    encoder_model = RobertaModel.from_pretrained(args.fine_tuned_model).to(device)
    faiss_index = faiss.read_index(os.path.join(args.faiss_dir, "populated.index"))
    faiss_index.nprobe = 64

    print(
        search_code_documents(args.query, encoder_model, query_tokenizer, n_results=50)
    )
