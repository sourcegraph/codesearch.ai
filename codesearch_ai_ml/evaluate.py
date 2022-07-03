import numpy as np
from scipy.spatial.distance import cdist

from codesearch_ai_ml.util import torch_gpu_to_np


def compute_mrr(query_embeddings, code_embeddings):
    distance_matrix = cdist(query_embeddings, code_embeddings, "cosine")
    correct_elements = np.expand_dims(np.diag(distance_matrix), axis=-1)
    ranks = np.sum(distance_matrix <= correct_elements, axis=-1)
    ranks = ranks[
        np.invert(np.isnan(ranks)) & (ranks >= 1)
    ]  # Make sure we only use valid ranks
    return float(np.mean(1.0 / ranks))


def evaluate_mrr(model, data_loader, batch_size, mrr_batch_size) -> float:
    mrrs = []
    batch_code_embeddings, batch_query_embeddings = [], []
    for batch in data_loader:
        code_outputs, query_outputs = model.encode_inputs(*batch)

        code_embeddings = torch_gpu_to_np(code_outputs)
        query_embeddings = torch_gpu_to_np(query_outputs)

        batch_code_embeddings.append(code_embeddings)
        batch_query_embeddings.append(query_embeddings)

        if len(batch_code_embeddings) == mrr_batch_size // batch_size:
            mrrs.append(
                compute_mrr(
                    np.vstack(batch_query_embeddings), np.vstack(batch_code_embeddings)
                )
            )
            batch_code_embeddings, batch_query_embeddings = [], []

    return np.mean(mrrs)
