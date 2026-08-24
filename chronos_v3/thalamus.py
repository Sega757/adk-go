import asyncio
from dataclasses import dataclass, field
from typing import Any

@dataclass(order=True)
class Event:
    """
    Represents an event in the Thalamus event bus.
    Priority determines execution order (lower value = higher priority).
    """
    priority: int
    name: str = field(compare=False)
    payload: Any = field(default=None, compare=False)

class ThalamusBus:
    """
    The Thalamus Event Bus.
    Acts as an asynchronous event bus built on a non-blocking queue structure.
    """
    def __init__(self):
        self._queue = asyncio.PriorityQueue()
        self._is_running = True

    async def publish(self, priority: int, name: str, payload: Any = None):
        """
        Publishes an event to the bus.
        """
        if self._is_running:
            event = Event(priority=priority, name=name, payload=payload)
            await self._queue.put(event)

    async def get(self) -> Event:
        """
        Retrieves the next highest priority event from the bus.
        """
        return await self._queue.get()

    def task_done(self):
        """
        Marks the previously retrieved event as processed.
        """
        self._queue.task_done()

    def halt(self):
        """
        Halts the bus, preventing new events from being published.
        Used when the Absolute Kill Switch (Panic Mode) is triggered.
        """
        self._is_running = False

    @property
    def qsize(self) -> int:
        """Returns the number of events currently in the queue."""
        return self._queue.qsize()
