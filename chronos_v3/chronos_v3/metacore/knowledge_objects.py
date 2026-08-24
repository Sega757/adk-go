import numpy as np


class KnowledgeObjectStore:
    def __init__(self, value_dim: int, key_dim: int):
        self.v_dim = value_dim
        self.k_dim = key_dim
        # Тензорная матрица памяти внешних произведений: M = sum(v_i (x) k_i^T)
        self.M = np.zeros((value_dim, key_dim), dtype=np.float64)

    def store_fact(self, key_vector: np.ndarray, value_vector: np.ndarray):
        k = key_vector.reshape(-1, 1)
        v = value_vector.reshape(-1, 1)

        # Нормализация
        k = k / (np.linalg.norm(k) + 1e-12)
        v = v / (np.linalg.norm(v) + 1e-12)

        outer_product = np.dot(v, k.T)
        self.M += outer_product

    def retrieve(self, query_key: np.ndarray) -> np.ndarray:
        k_q = query_key.reshape(-1, 1)
        k_q = k_q / (np.linalg.norm(k_q) + 1e-12)
        retrieved_v = np.dot(self.M, k_q)
        return retrieved_v.flatten()
