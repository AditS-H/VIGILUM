"""Inference module for VIGILUM ML."""

from vigilum_ml.inference.service import (
    InferenceService,
    export_to_onnx,
)

__all__ = ["InferenceService", "export_to_onnx"]
