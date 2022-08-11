import os
import json

import faiss
import torch
from fastapi import FastAPI
from transformers import RobertaModel
from codesearch_ai_ml.search import search_code_documents

from codesearch_ai_ml.train_tokenizer import get_text_tokenizer, get_code_tokenizer


def get_faiss_index(path):
    faiss_index = faiss.read_index(os.path.join(path, "populated.index"))
    faiss_index.nprobe = 64
    return faiss_index


def read_json_file(path):
    with open(path, encoding="utf-8") as f:
        return json.load(f)


app = FastAPI()

device = torch.device("cuda")


query_tokenizer = get_text_tokenizer(os.environ.get("BASE_MODEL"))
code_tokenizer = get_code_tokenizer(os.environ.get("BASE_MODEL"))
extracted_functions_encoder_model = RobertaModel.from_pretrained(
    os.environ.get("EXTRACTED_FUNCTIONS_FINE_TUNED_MODEL"), local_files_only=True
).to(device)
so_encoder_model = RobertaModel.from_pretrained(
    os.environ.get("SO_FINE_TUNED_MODEL"), local_files_only=True
).to(device)

extracted_functions_faiss_index = get_faiss_index(
    os.environ.get("EXTRACTED_FUNCTIONS_FAISS_DIR")
)
so_faiss_index = get_faiss_index(os.environ.get("SO_FAISS_DIR"))


extracted_functions_id_mapping = read_json_file(
    os.environ.get("EXTRACTED_FUNCTIONS_ID_MAPPING")
)
so_id_mapping = read_json_file(os.environ.get("SO_ID_MAPPING"))


@app.get("/search/functions/by-text")
def search_extracted_functions_by_text(query: str = "", count: int = 30):
    extracted_functions_indices = search_code_documents(
        query,
        extracted_functions_encoder_model,
        extracted_functions_faiss_index,
        query_tokenizer,
        n_results=count,
    )

    return {
        "ids": [
            extracted_functions_id_mapping[index]
            for index in extracted_functions_indices
        ],
    }


@app.get("/search/so/by-text")
def search_so_by_text(query: str = "", count: int = 30):
    questions_indices = search_code_documents(
        query, so_encoder_model, so_faiss_index, query_tokenizer, n_results=count
    )
    return {
        "ids": [so_id_mapping[index] for index in questions_indices],
    }


@app.get("/search/functions/by-code")
def search_functions_by_code(query: str = "", count: int = 30):
    extracted_functions_indices = search_code_documents(
        query,
        extracted_functions_encoder_model,
        extracted_functions_faiss_index,
        code_tokenizer,
        n_results=count,
    )

    return {
        "ids": [
            extracted_functions_id_mapping[index]
            for index in extracted_functions_indices
        ],
    }


@app.get("/search/so/by-code")
def search_so_by_code(query: str = "", count: int = 30):
    questions_indices = search_code_documents(
        query, so_encoder_model, so_faiss_index, code_tokenizer, n_results=count
    )

    return {
        "ids": [so_id_mapping[index] for index in questions_indices],
    }
