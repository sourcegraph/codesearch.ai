import time
import itertools
import logging
import datetime
import argparse

import torch
from torch import nn
from torch.utils.data import DataLoader
from torch.nn.functional import cross_entropy
from transformers import DataCollatorWithPadding, RobertaModel, get_scheduler
from torch.optim import AdamW
from datasets import load_from_disk
import numpy as np

from codesearch_ai_ml.evaluate import evaluate_mrr
from codesearch_ai_ml.model_config import SMALL_NUMBER
from codesearch_ai_ml.train_tokenizer import get_code_tokenizer


device = torch.device("cuda") if torch.cuda.is_available() else torch.device("cpu")


class CodeSearchFineTuningModel(nn.Module):
    def __init__(self, encoder_model):
        super(CodeSearchFineTuningModel, self).__init__()
        self.encoder_model = encoder_model
        self.temp = nn.parameter.Parameter(torch.tensor(0.0), requires_grad=True)

    def encode_inputs(self, code_inputs, query_inputs):
        code_encoded = self.encoder_model(**code_inputs).pooler_output
        queries_encoded = self.encoder_model(**query_inputs).pooler_output
        return code_encoded, queries_encoded

    def forward(self, code_inputs, query_inputs):
        # Encode inputs
        code_encoded, queries_encoded = self.encode_inputs(code_inputs, query_inputs)

        # Cosine similarity
        code_encoded_norm = torch.norm(code_encoded, dim=1, keepdim=True) + SMALL_NUMBER
        queries_encoded_norm = (
            torch.norm(queries_encoded, dim=1, keepdim=True) + SMALL_NUMBER
        )

        return (
            (queries_encoded / queries_encoded_norm) * torch.exp(self.temp),
            (code_encoded / code_encoded_norm) * torch.exp(self.temp),
        )


def fine_tune_model(
    base_model_path: str,
    fine_tuned_model_path: str,
    train_pairs_dataset_loader,
    test_pairs_dataset_loader,
    num_epochs: int,
    batch_size: int,
    learning_rate: float,
    warmup_ratio: float,
):
    encoder_model = RobertaModel.from_pretrained(base_model_path)
    encoder_model.to(device)

    fine_tuning_model = CodeSearchFineTuningModel(encoder_model)
    fine_tuning_model = nn.DataParallel(fine_tuning_model)
    fine_tuning_model.to(device)

    optimizer = AdamW(fine_tuning_model.parameters(), lr=learning_rate)
    num_training_steps = num_epochs * len(train_pairs_dataset_loader)
    lr_scheduler = get_scheduler(
        "linear",  # TODO: try different schedulers
        optimizer=optimizer,
        num_warmup_steps=int(warmup_ratio * num_training_steps),
        num_training_steps=num_training_steps,
    )

    for epoch in range(num_epochs):
        logging.info(f"=== Epoch {epoch + 1} ===")
        epoch_start = time.time()

        fine_tuning_model.train()
        step = 0
        losses = []
        for batch in train_pairs_dataset_loader:
            # Encode code and queries
            queries_encoded, code_encoded = fine_tuning_model(*batch)
            similarity_matrix = torch.matmul(
                queries_encoded,
                code_encoded.t(),
            )

            # Compute loss
            labels = torch.arange(similarity_matrix.shape[0], device=device)
            row_loss = cross_entropy(similarity_matrix, labels)
            col_loss = cross_entropy(similarity_matrix.t(), labels)
            loss = (row_loss + col_loss) / 2.0
            loss.backward()
            losses.append(loss.item())

            # Update weights
            optimizer.step()
            lr_scheduler.step()
            optimizer.zero_grad()

            if step % 500 == 0:
                logging.info(
                    f"{step}/{len(train_pairs_dataset_loader)}: {np.mean(losses)}"
                )

            step += 1

        epoch_duration = time.time() - epoch_start
        logging.info(f"Mean loss: {np.mean(losses)}, duration: {epoch_duration:.1f}s")

        fine_tuning_model.eval()
        logging.info("Eval")
        with torch.no_grad():
            val_start = time.time()
            validation_mean_mrr = evaluate_mrr(
                fine_tuning_model.module, test_pairs_dataset_loader, batch_size, 1024
            )
            val_duration = time.time() - val_start
            logging.info(f"MRR: {validation_mean_mrr}, duration: {val_duration:.1f}s")

        logging.info("Saving model...")
        encoder_model.save_pretrained(fine_tuned_model_path)


def keys_to_device(d):
    return {k: v.to(device) for k, v in d.items()}


def collate_fn(batch_collator):
    def _collate_fn(batch):
        code_batch = batch_collator(
            [
                {
                    "input_ids": sample["code_input_ids"],
                    "attention_mask": sample["code_attention_mask"],
                }
                for sample in batch
            ]
        )
        query_batch = batch_collator(
            [
                {
                    "input_ids": sample["query_input_ids"],
                    "attention_mask": sample["query_attention_mask"],
                }
                for sample in batch
            ]
        )
        return keys_to_device(code_batch), keys_to_device(query_batch)

    return _collate_fn


def main(
    base_model_path: str,
    fine_tuned_model_path: str,
    train_dataset_path: str,
    test_dataset_path: str,
):
    num_epochs = [5]
    batch_size = [256]
    learning_rate = [5e-5]
    warmup_ratio = [0.15]

    for e, b, l, w in itertools.product(
        num_epochs, batch_size, learning_rate, warmup_ratio
    ):
        logging.info(dict(num_epochs=e, batch_size=b, learning_rate=l, warmup_ratio=w))

        code_tokenizer = get_code_tokenizer(base_model_path)
        train_dataset = load_from_disk(train_dataset_path, keep_in_memory=True)
        train_dataset.set_format(
            "torch",
            columns=[
                "code_input_ids",
                "code_attention_mask",
                "query_input_ids",
                "query_attention_mask",
            ],
        )

        test_dataset = load_from_disk(test_dataset_path, keep_in_memory=True)
        test_dataset.set_format(
            "torch",
            columns=[
                "code_input_ids",
                "code_attention_mask",
                "query_input_ids",
                "query_attention_mask",
            ],
        )

        batch_collator = DataCollatorWithPadding(tokenizer=code_tokenizer)

        train_pairs_dataset_loader = DataLoader(
            train_dataset,
            batch_size=b,
            shuffle=True,
            drop_last=True,
            collate_fn=collate_fn(batch_collator),
        )
        test_pairs_dataset_loader = DataLoader(
            test_dataset,
            batch_size=b,
            shuffle=True,
            drop_last=True,
            collate_fn=collate_fn(batch_collator),
        )

        fine_tune_model(
            base_model_path,
            fine_tuned_model_path,
            train_pairs_dataset_loader,
            test_pairs_dataset_loader,
            num_epochs=e,
            batch_size=b,
            learning_rate=l,
            warmup_ratio=w,
        )


if __name__ == "__main__":
    logging.basicConfig(
        filename="train_log_"
        + datetime.datetime.now().strftime("%Y_%m_%d_%H_%M_%S")
        + ".log",
        level=logging.INFO,
        format="%(asctime)s - %(message)s",
        datefmt="%d/%m/%Y %I:%M:%S",
    )

    parser = argparse.ArgumentParser()
    parser.add_argument("--base-model", dest="base_model")
    parser.add_argument("--fine-tuned-model", dest="fine_tuned_model")
    parser.add_argument("--train-dataset", dest="train_dataset")
    parser.add_argument("--test-dataset", dest="test_dataset")
    args, _ = parser.parse_known_args()

    main(args.base_model, args.fine_tuned_model, args.train_dataset, args.test_dataset)
