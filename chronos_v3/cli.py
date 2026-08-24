import asyncio
import numpy as np
from chronos_v3.thalamus import ThalamusBus
from chronos_v3.metacore import MetaCoreShield

async def prefrontal_cortex_worker(bus: ThalamusBus, shield: MetaCoreShield, worker_id: int):
    print(f"[Prefrontal Cortex-{worker_id}] Starting predictive semantic cache prefetching...")
    while shield.panic_mode is False:
        try:
            event = await asyncio.wait_for(bus.get(), timeout=0.2)

            # Allow Cingulate Cortex to process the anomaly
            if event.name == "InjectAnomaly":
                # Put it back and yield so Cingulate Cortex can grab it
                await bus.publish(event.priority, event.name, event.payload)
                bus.task_done()
                await asyncio.sleep(0.01)
                continue

            print(f"[Prefrontal Cortex-{worker_id}] Processing Event: {event.name} (Priority: {event.priority})")
            await asyncio.sleep(0.05)

            if "semantic" in event.name.lower():
                print(f"[Prefrontal Cortex-{worker_id}] Evaluating semantic density...")
                embeddings = np.random.randn(10, 3)
                if shield.check_stability(embeddings):
                    print(f"!!! [Prefrontal Cortex-{worker_id}] METACORE KILLED EXECUTION: Semantic density too high!")
                    bus.halt()

            bus.task_done()
        except asyncio.TimeoutError:
            continue
    print(f"[Prefrontal Cortex-{worker_id}] Halted due to Panic Mode.")

async def cingulate_cortex_worker(bus: ThalamusBus, shield: MetaCoreShield):
    print("[Cingulate Cortex] Starting context-state router...")
    while shield.panic_mode is False:
        try:
            event = await asyncio.wait_for(bus.get(), timeout=0.2)

            if event.name != "InjectAnomaly" and "semantic" not in event.name.lower():
                 print(f"[Cingulate Cortex] Routing State Transition: {event.name} (Priority: {event.priority})")
            elif event.name == "InjectAnomaly":
                 pass # process below
            else:
                 # Yield back semantic events to PFC
                 await bus.publish(event.priority, event.name, event.payload)
                 bus.task_done()
                 await asyncio.sleep(0.01)
                 continue

            await asyncio.sleep(0.05)

            if event.name == "InjectAnomaly":
                print(f"[Cingulate Cortex] Routing State Transition: {event.name} (Priority: {event.priority})")
                print("[Cingulate Cortex] Warning: Anomaly detected, evaluating...")
                embeddings = np.array([
                    [1.0, 0.05, 0.01],
                    [0.99, 0.0, -0.05],
                    [1.0, -0.02, 0.03],
                    [0.98, 0.01, 0.0]
                ])
                if shield.check_stability(embeddings):
                    print("!!! [Cingulate Cortex] METACORE TRIGGERED ABSOLUTE KILL SWITCH (Panic Mode) !!!")
                    bus.halt()

            bus.task_done()
        except asyncio.TimeoutError:
            continue
    print("[Cingulate Cortex] Halted due to Panic Mode.")

async def simulate_traffic(bus: ThalamusBus, shield: MetaCoreShield):
    print("[Traffic Simulator] Starting...")

    for i in range(3):
        if shield.panic_mode: break
        await bus.publish(priority=10, name=f"SemanticPrefetch_Batch_{i}")
        await asyncio.sleep(0.05)

    for i in range(2):
        if shield.panic_mode: break
        await bus.publish(priority=2, name=f"UI_Interaction_{i}")
        await asyncio.sleep(0.05)

    if not shield.panic_mode:
        print("[Traffic Simulator] Injecting critical anomaly event...")
        await bus.publish(priority=0, name="InjectAnomaly")

    await asyncio.sleep(0.2)
    print("[Traffic Simulator] Attempting to send post-anomaly payload...")
    await bus.publish(priority=1, name="MaliciousPayload")

    # Let workers process
    await asyncio.sleep(0.5)
    print("[Traffic Simulator] Finished.")

async def main():
    print("=== Chronos V3 Core Architecture Simulation ===")
    bus = ThalamusBus()
    shield = MetaCoreShield(threshold=0.6)

    pfc_task = asyncio.create_task(prefrontal_cortex_worker(bus, shield, 1))
    cc_task = asyncio.create_task(cingulate_cortex_worker(bus, shield))
    traffic_task = asyncio.create_task(simulate_traffic(bus, shield))

    try:
        await asyncio.wait_for(asyncio.gather(pfc_task, cc_task, traffic_task), timeout=5.0)
    except asyncio.TimeoutError:
        print("Simulation timeout reached.")

    print("=== Simulation Complete ===")
    print(f"Final Panic Mode State: {shield.panic_mode}")
    print(f"Events remaining in queue: {bus.qsize}")

if __name__ == "__main__":
    asyncio.run(main())
