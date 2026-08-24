import logging
from enum import Enum
from ..core.events import Event, Priority
from ..core.thalamus import ThalamusEventBus

logger = logging.getLogger("CingulateCortex")


class SystemState(str, Enum):
    ANALYTICS = "ANALYTICS"
    OBSERVATION = "OBSERVATION"
    REFLECTION = "REFLECTION"


class CingulateCortex:
    def __init__(self, bus: ThalamusEventBus):
        self.bus = bus
        self.current_state = SystemState.OBSERVATION

    async def handle_context_signal(self, event: Event):
        requested_mode = event.payload.get("requested_state")
        if requested_mode and requested_mode in SystemState.__members__:
            old_state = self.current_state
            self.current_state = SystemState(requested_mode)
            logger.info(f"State transition: {old_state.value} -> {self.current_state.value}")

            notification = Event(
                priority=Priority.HIGH_CONTROL,
                event_type="STATE_TRANSITIONED",
                payload={"from": old_state.value, "to": self.current_state.value},
                source="CingulateCortex"
            )
            await self.bus.publish(notification)
