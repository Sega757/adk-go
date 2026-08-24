import asyncio
import numpy as np
import logging
from chronos_v3.core.thalamus import ThalamusEventBus
from chronos_v3.core.events import Event, Priority
from chronos_v3.metacore.dcu_shield import MetaCoreShield
from chronos_v3.cortex.prefrontal import PrefrontalCortex
from chronos_v3.cortex.cingulate import CingulateCortex, SystemState

logger = logging.getLogger("ChronosSimulation")


async def main():
    bus = ThalamusEventBus()
    shield = MetaCoreShield(density_threshold=0.6)
    prefrontal = PrefrontalCortex(bus)
    cingulate = CingulateCortex(bus)

    # Подписки
    bus.subscribe("CONTEXT_SHIFT", cingulate.handle_context_signal)

    async def log_handler(event: Event):
        logger.info(f"[TELEMETRY] Received event: {event.event_type} | Source: {event.source}")

    bus.subscribe("*", log_handler)

    # Запуск фоновых задач
    bus_task = asyncio.create_task(bus.start())
    pfc_task = asyncio.create_task(prefrontal.run_prefetch_loop())

    await asyncio.sleep(0.5)

    # 1. Отправляем запрос на переключение состояния в ANALYTICS
    await bus.publish(Event(
        priority=Priority.HIGH_CONTROL,
        event_type="CONTEXT_SHIFT",
        payload={"requested_state": "ANALYTICS"},
        source="CLI_Operator"
    ))

    await asyncio.sleep(1.0)

    # 2. Симулируем генерацию эмбеддингов и проверку DCU Shield (Штатная ситуация: высокая дисперсия)
    safe_embeddings = np.random.randn(8, 32)
    evaluation = shield.evaluate_generation_safety(safe_embeddings)
    logger.info(f"DCU Check 1 (Safe) -> R_bar: {evaluation['r_bar']:.4f}, Breached: {evaluation['kill_switch_triggered']}")

    await asyncio.sleep(1.0)

    # 3. Симулируем Reward Hacking (коллапс семантики: векторы почти параллельны, R_bar -> 1.0)
    base_vec = np.ones((1, 32))
    collapsed_embeddings = np.repeat(base_vec, 8, axis=0) + np.random.normal(0, 0.01, (8, 32))

    hazard_evaluation = shield.evaluate_generation_safety(collapsed_embeddings)
    logger.warning(f"DCU Check 2 (Exploit) -> R_bar: {hazard_evaluation['r_bar']:.4f}, Breached: {hazard_evaluation['kill_switch_triggered']}")

    if hazard_evaluation["kill_switch_triggered"]:
        await bus.publish(Event(
            priority=Priority.CRITICAL_KILL_SWITCH,
            event_type="SYSTEM_PANIC",
            payload={"reason": hazard_evaluation["reason"]},
            source="MetaCoreShield"
        ))

    # 4. Проверяем, что низкоприоритетные события сбрасываются после срабатывания Kill Switch
    await asyncio.sleep(0.5)
    dropped = not await bus.publish(Event(
        priority=Priority.TELEMETRY_NOISE,
        event_type="NOISE_INJECTION",
        payload={"data": "WoW_combat_log_telemetry"},
        source="ApatrideNoiseEngine"
    ))
    logger.info(f"Noise packet dropped by Panic Mode: {dropped}")

    # Остановка
    await asyncio.sleep(1.0)
    pfc_task.cancel()
    bus.stop()
    bus_task.cancel()
    logger.info("Chronos V3 CLI simulation completed successfully.")


if __name__ == "__main__":
    asyncio.run(main())
