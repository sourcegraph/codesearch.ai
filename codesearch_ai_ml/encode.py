import torch

from codesearch_ai_ml.model_config import SMALL_NUMBER
from codesearch_ai_ml.util import torch_gpu_to_np
from codesearch_ai_ml.train_tokenizer import tokenize


def encode_code_samples(encoder_model, code_tokenizer, samples, device):
    encodings = tokenize(
        code_tokenizer, samples, padding=True, return_special_tokens_mask=False
    ).to(device)

    with torch.no_grad():
        output = encoder_model(**encodings).pooler_output
        output_norm = torch.norm(output, dim=1, keepdim=True) + SMALL_NUMBER
        return torch_gpu_to_np(output / output_norm)
