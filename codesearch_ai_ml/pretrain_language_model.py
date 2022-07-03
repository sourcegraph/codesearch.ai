import argparse

from datasets import load_from_disk
from transformers import (
    Trainer,
    TrainingArguments,
    RobertaConfig,
    RobertaForMaskedLM,
    DataCollatorForLanguageModeling,
)

from codesearch_ai_ml.model_config import (
    MAX_LENGTH,
    NUM_ATTENTION_HEADS,
    NUM_HIDDEN_LAYERS,
    HIDDEN_SIZE,
)
from codesearch_ai_ml.train_tokenizer import get_code_tokenizer


def pretrain_mlm_model(model_path, tokenizer, train_dataset, test_dataset):
    config = RobertaConfig(
        vocab_size=len(tokenizer),
        max_position_embeddings=(MAX_LENGTH * 2),  # Something large, just in case
        num_attention_heads=NUM_ATTENTION_HEADS,
        num_hidden_layers=NUM_HIDDEN_LAYERS,
        type_vocab_size=1,
        hidden_size=HIDDEN_SIZE,
    )

    model = RobertaForMaskedLM(config=config)

    training_args = TrainingArguments(
        output_dir=model_path,
        overwrite_output_dir=True,
        do_train=True,
        do_eval=True,
        num_train_epochs=3,
        warmup_ratio=0.1,
        per_device_train_batch_size=32,
        per_device_eval_batch_size=32,
        logging_strategy="steps",
        logging_steps=1000,
        save_strategy="steps",
        save_steps=100_000,
        evaluation_strategy="steps",
        eval_steps=100_000,
        save_total_limit=1,  # TODO: load_best_model_at_end only considers the last saved checkpoint
        prediction_loss_only=True,
        load_best_model_at_end=True,
        disable_tqdm=False,
        # Speed optimizations
        group_by_length=True,
        length_column_name="length",
        fp16=True,
        ddp_find_unused_parameters=False,
    )

    data_collator = DataCollatorForLanguageModeling(
        tokenizer=tokenizer, mlm=True, mlm_probability=0.15
    )

    trainer = Trainer(
        model=model,
        args=training_args,
        data_collator=data_collator,
        train_dataset=train_dataset,
        eval_dataset=test_dataset,
    )

    # TODO: resume_from_checkpoint arg
    trainer.train(resume_from_checkpoint=False)
    trainer.save_model(model_path)


def pretrain_model(model_path: str, train_dataset_path: str, test_dataset_path: str):
    code_tokenizer = get_code_tokenizer(model_path)

    train_dataset = load_from_disk(train_dataset_path)
    train_dataset.set_format(
        type="pt",
        columns=["input_ids", "special_tokens_mask", "attention_mask"],
    )

    test_dataset = load_from_disk(test_dataset_path)
    test_dataset.set_format(
        type="pt",
        columns=["input_ids", "special_tokens_mask", "attention_mask"],
    )

    # We need a tokenizer for the DataCollatorForLanguageModeling to get the mask/pad token ids.
    # Either code or query tokenizer is valid since they share both of those tokens.
    pretrain_mlm_model(model_path, code_tokenizer, train_dataset, test_dataset)


# To launch the script: python -m torch.distributed.launch --nproc_per_node=4
if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--model-path", dest="model_path")
    parser.add_argument("--train-dataset", dest="train_dataset")
    parser.add_argument("--test-dataset", dest="test_dataset")
    args, _ = parser.parse_known_args()

    pretrain_model(args.model_path, args.train_dataset, args.test_dataset)
