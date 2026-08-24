import asyncio
import pytest
from chronos_v3.thalamus import ThalamusBus, Event

@pytest.mark.asyncio
async def test_thalamus_priority_routing():
    bus = ThalamusBus()

    # Publish events out of order
    await bus.publish(priority=10, name="LowPriorityTelemetry")
    await bus.publish(priority=1, name="CriticalSystemEvent")
    await bus.publish(priority=5, name="NormalOperation")

    # Retrieve events and assert priority order
    event1 = await bus.get()
    assert event1.priority == 1
    assert event1.name == "CriticalSystemEvent"
    bus.task_done()

    event2 = await bus.get()
    assert event2.priority == 5
    assert event2.name == "NormalOperation"
    bus.task_done()

    event3 = await bus.get()
    assert event3.priority == 10
    assert event3.name == "LowPriorityTelemetry"
    bus.task_done()

    assert bus.qsize == 0

@pytest.mark.asyncio
async def test_thalamus_halt():
    bus = ThalamusBus()

    await bus.publish(1, "BeforeHalt")
    assert bus.qsize == 1

    bus.halt()

    await bus.publish(2, "AfterHalt")
    assert bus.qsize == 1 # Queue size should not increase after halt

    event = await bus.get()
    assert event.name == "BeforeHalt"
    bus.task_done()

    assert bus.qsize == 0
