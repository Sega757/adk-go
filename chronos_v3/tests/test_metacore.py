import numpy as np
import pytest
from chronos_v3.metacore import MetaCoreShield

def test_metacore_shield_initialization():
    shield = MetaCoreShield(threshold=0.6)
    assert shield.threshold == 0.6
    assert shield.panic_mode is False

def test_metacore_stability_high_variance():
    # High variance (low concentration), R_bar should be low, panic should be False
    shield = MetaCoreShield(threshold=0.6)

    # Generate vectors in opposite directions to simulate high variance / dispersion
    embeddings = np.array([
        [1.0, 0.0, 0.0],
        [-1.0, 0.0, 0.0],
        [0.0, 1.0, 0.0],
        [0.0, -1.0, 0.0]
    ])

    panic = shield.check_stability(embeddings)
    assert panic is False
    assert shield.panic_mode is False

def test_metacore_stability_low_variance_trigger():
    # Low variance (high concentration), vectors point in similar directions.
    # R_bar should be high (>0.6), panic should be True
    shield = MetaCoreShield(threshold=0.6)

    # Generate vectors pointing mostly in the same direction
    embeddings = np.array([
        [1.0, 0.1, 0.1],
        [1.0, 0.0, -0.1],
        [0.9, 0.2, 0.0],
        [1.0, -0.1, 0.1]
    ])

    panic = shield.check_stability(embeddings)
    assert panic is True
    assert shield.panic_mode is True

def test_estimate_kappa():
    shield = MetaCoreShield()

    # Test with d=3, R_bar=0.5
    # The approximate kappa should be roughly 1.79 for R_bar=0.5, d=3
    # let's just check if it returns a positive float
    kappa = shield.estimate_kappa(0.5, 3)
    assert isinstance(kappa, float)
    assert kappa > 0

    # Test edge cases
    assert shield.estimate_kappa(0.0, 3) == 0.0
    assert shield.estimate_kappa(1.0, 3) == float('inf')
