import asyncio
import logging
from typing import Callable, Coroutine, Dict, List, Any
from .events import Event, Priority

logging.basicConfig(level=logging.INFO, format="[%(asctime)s] [%(levelname)s] [%(name)s]: %(message)s")
logger = logging.getLogger("Thalamus")


class ThalamusEventBus:
    def __init__(self):
        self._queue: asyncio.PriorityQueue[Event] = asyncio.PriorityQueue()
        self._subscribers: Dict[str, List[Callable[[Event], Coroutine[Any, Any, None]]]] = {}
        self.panic_mode: bool = False
        self._running: bool = False

    def subscribe(self, event_type: str, handler: Callable[[Event], Coroutine[Any, Any, None]]):
        if event_type not in self._subscribers:
            self._subscribers[event_type] = []
        self._subscribers[event_type].append(handler)

    async def publish(self, event: Event) -> bool:
        # Синхронный фильтр Panic Mode перед постановкой в очередь
        if self.panic_mode and event.priority > Priority.CRITICAL_KILL_SWITCH:
            logger.warning(f"Panic Mode ACTIVE. Dropping event: {event.event_type} from {event.source}")
            return False

        await self._queue.put(event)
        return True

    def trip_kill_switch(self, reason: str):
        self.panic_mode = True
        logger.critical(f"KILL SWITCH TRIGGERED. Panic Mode enabled. Reason: {reason}")

    def reset_kill_switch(self):
        self.panic_mode = False
        logger.info("Panic Mode CLEARED. Normal signal flow resumed.")

    async def start(self):
        self._running = True
        logger.info("Thalamus Event Bus metronome started.")
        while self._running:
            try:
                event = await self._queue.get()

                if event.event_type == "SYSTEM_PANIC":
                    self.trip_kill_switch(event.payload.get("reason", "Unknown panic signal"))
                    self._queue.task_done()
                    continue

                if self.panic_mode and event.priority > Priority.CRITICAL_KILL_SWITCH:
                    self._queue.task_done()
                    continue

                handlers = self._subscribers.get(event.event_type, [])
                handlers.extend(self._subscribers.get("*", []))

                for handler in handlers:
                    asyncio.create_task(handler(event))

                self._queue.task_done()
            except asyncio.CancelledError:
                break

    def stop(self):
        self._running = False
