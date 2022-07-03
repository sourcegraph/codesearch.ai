import itertools

import torch


def torch_gpu_to_np(tensor: torch.Tensor):
    return tensor.cpu().numpy()


def extract_first_tensor(encodings):
    return {key: val[0].clone().detach() for key, val in encodings.items()}


def flatten(iterable):
    return itertools.chain.from_iterable(iterable)


def batches(stream, n):
    items = []
    for item in stream:
        items.append(item)
        if len(items) == n:
            yield items
            items = []

    if len(items) > 0:
        yield items
