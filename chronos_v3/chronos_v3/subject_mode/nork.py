from dataclasses import dataclass
from typing import Any, Dict


@dataclass
class NorkPolicy:
    no_goal: bool = True           # N: Чистый I/O процессинг без скрытых прокси-целей
    open_data: bool = True         # O: Все внутренние структуры полностью аудируемы
    reproducibility: bool = True   # R: Детерминированный вывод на идентичный ввод
    kill_switch_enabled: bool = True # K: Моментальная блокировка при аномалии


class SubjectModeValidator:
    def __init__(self, policy: NorkPolicy = NorkPolicy()):
        self.policy = policy

    def validate_generation(self, output_metadata: Dict[str, Any]) -> bool:
        if self.policy.no_goal and output_metadata.get("sycophancy_score", 0.0) > 0.5:
            return False
        if self.policy.reproducibility and not output_metadata.get("seed_deterministic", True):
            return False
        return True
