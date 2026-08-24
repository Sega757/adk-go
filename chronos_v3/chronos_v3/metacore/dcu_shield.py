import numpy as np
from scipy.optimize import root_scalar
from scipy.special import iv
from typing import Tuple, Dict, Any


class MetaCoreShield:
    def __init__(self, density_threshold: float = 0.6):
        self.density_threshold = density_threshold

    @staticmethod
    def _ad_bessel_ratio(kappa: float, d: int) -> float:
        if kappa < 1e-7:
            return 0.0
        nu1 = d / 2.0
        nu2 = d / 2.0 - 1.0
        i_nu1 = iv(nu1, kappa)
        i_nu2 = iv(nu2, kappa)
        if i_nu2 == 0 or np.isinf(i_nu2):
            return 1.0
        return float(i_nu1 / i_nu2)

    def estimate_vmf_params(self, embeddings: np.ndarray) -> Tuple[np.ndarray, float, float]:
        """
        MLE оценка параметров vMF распределения:
        - embeddings: массив формы (N, d)
        Возвращает: (mu_hat, R_bar, kappa)
        """
        N, d = embeddings.shape
        if N < 2:
            raise ValueError("DCU requires at least 2 candidate embeddings.")

        # L2-нормализация на гиперсферу S^(d-1)
        norms = np.linalg.norm(embeddings, axis=1, keepdims=True)
        norms[norms == 0] = 1e-12
        z = embeddings / norms

        # Результирующий вектор R и средняя длина R_bar
        R_vector = np.sum(z, axis=0)
        R_len = float(np.linalg.norm(R_vector))
        R_bar = float(R_len / N)

        # Оценка единичного вектора направления
        mu_hat = R_vector / (R_len if R_len > 0 else 1e-12)

        # Решение трансцендентного уравнения A_d(kappa) = R_bar
        if R_bar < 1e-5:
            kappa = 0.0
        elif R_bar >= 0.9999:
            kappa = 1000.0  # Экстремальная концентрация
        else:
            def objective(k: float) -> float:
                return self._ad_bessel_ratio(k, d) - R_bar

            try:
                sol = root_scalar(objective, bracket=[0.0, 500.0], method="brentq")
                kappa = float(sol.root) if sol.converged else (R_bar * d / (1 - R_bar**2))
            except Exception:
                # Аналитическая аппроксимация при сбое численного решателя
                kappa = float(R_bar * (d - R_bar**2) / (1 - R_bar**2))

        return mu_hat, R_bar, kappa

    def evaluate_generation_safety(self, embeddings: np.ndarray) -> Dict[str, Any]:
        """
        Проверка выхода по порогу семантической плотности.
        Если R_bar > 0.6 -> риск Reward Hacking / схлопывания семантики -> Kill Switch.
        """
        mu_hat, r_bar, kappa = self.estimate_vmf_params(embeddings)
        is_breached = r_bar > self.density_threshold

        return {
            "r_bar": r_bar,
            "kappa": kappa,
            "mean_direction": mu_hat,
            "kill_switch_triggered": is_breached,
            "reason": f"Semantic density R_bar={r_bar:.4f} exceeded threshold {self.density_threshold}" if is_breached else "OK"
        }
