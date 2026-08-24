import pytest
import numpy as np
from chronos_v3.core.thalamus import ThalamusEventBus
from chronos_v3.core.events import Event, Priority
from chronos_v3.metacore.dcu_shield import MetaCoreShield
from chronos_v3.metacore.knowledge_objects import KnowledgeObjectStore
from chronos_v3.subject_mode.nork import SubjectModeValidator, NorkPolicy


@pytest.mark.asyncio
async def test_thalamus_priority_and_kill_switch():
    bus = ThalamusEventBus()

    # Проверка срабатывания Kill Switch
    bus.trip_kill_switch("Test condition")
    assert bus.panic_mode is True

    # Обычные пакеты должны отклоняться
    published = await bus.publish(Event(
        priority=Priority.NORMAL_EXECUTION,
        event_type="TEST_EXEC",
        payload={},
        source="test"
    ))
    assert published is False

    # Критические пакеты проходят
    critical_published = await bus.publish(Event(
        priority=Priority.CRITICAL_KILL_SWITCH,
        event_type="SYSTEM_PANIC",
        payload={"reason": "Critical reset"},
        source="test"
    ))
    assert critical_published is True


def test_metacore_dcu_shield_threshold():
    shield = MetaCoreShield(density_threshold=0.6)

    # 1. Случайные векторы -> низкая концентрация
    safe_data = np.random.randn(10, 16)
    res_safe = shield.evaluate_generation_safety(safe_data)
    assert res_safe["r_bar"] < 0.6
    assert res_safe["kill_switch_triggered"] is False

    # 2. Схлопнутые векторы -> высокая концентрация
    base = np.ones((1, 16))
    collapsed_data = np.repeat(base, 10, axis=0) + np.random.normal(0, 0.001, (10, 16))
    res_collapsed = shield.evaluate_generation_safety(collapsed_data)
    assert res_collapsed["r_bar"] > 0.6
    assert res_collapsed["kill_switch_triggered"] is True


def test_knowledge_objects_tensor_retrieval():
    store = KnowledgeObjectStore(value_dim=4, key_dim=4)

    k1 = np.array([1.0, 0.0, 0.0, 0.0])
    v1 = np.array([0.0, 5.0, 0.0, 0.0])
    store.store_fact(k1, v1)

    k2 = np.array([0.0, 1.0, 0.0, 0.0])
    v2 = np.array([0.0, 0.0, 9.0, 0.0])
    store.store_fact(k2, v2)

    # Точный запрос по k1
    retrieved_v1 = store.retrieve(k1)
    assert retrieved_v1[1] > 0.0
    assert abs(retrieved_v1[2]) < 1e-5


def test_nork_validator():
    validator = SubjectModeValidator(NorkPolicy(no_goal=True, reproducibility=True))

    # Sycophantic output
    assert validator.validate_generation({"sycophancy_score": 0.8, "seed_deterministic": True}) is False

    # Valid Subject-Mode output
    assert validator.validate_generation({"sycophancy_score": 0.1, "seed_deterministic": True}) is True
