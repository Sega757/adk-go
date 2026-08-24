import time
from dataclasses import dataclass, field
from enum import IntEnum
from typing import Any, Dict


class Priority(IntEnum):
    CRITICAL_KILL_SWITCH = 0
    HIGH_CONTROL = 1
    NORMAL_EXECUTION = 2
    TELEMETRY_NOISE = 3


@dataclass(order=True)
class Event:
    priority: int
    timestamp: float = field(default_factory=time.time)
    event_type: str = field(compare=False, default="")
    payload: Dict[str, Any] = field(compare=False, default_factory=dict)
    source: str = field(compare=False, default="core")
