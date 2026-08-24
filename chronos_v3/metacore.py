import numpy as np
import scipy.special

class MetaCoreShield:
    def __init__(self, threshold: float = 0.6):
        """
        Initializes the MetaCore Mathematical Shield.

        Args:
            threshold (float): The critical stability threshold. If the mean resultant length
                               (rho / R_bar) exceeds this, it indicates semantic density
                               is too high (collapsing directions) and the kill switch
                               is triggered.
        """
        self.threshold = threshold
        self.panic_mode = False

    def check_stability(self, embeddings: np.ndarray) -> bool:
        """
        Evaluates a set of high-temperature candidate embeddings.

        Args:
            embeddings (np.ndarray): A 2D array of shape (N, d) representing N candidate
                                     outputs in d-dimensional space.

        Returns:
            bool: True if panic mode is triggered (i.e. stability check failed), False otherwise.
        """
        if len(embeddings) == 0:
            return False

        # Normalize the embeddings onto the unit hypersphere (L2 normalization)
        norms = np.linalg.norm(embeddings, axis=1, keepdims=True)
        # Avoid division by zero
        norms[norms == 0] = 1.0
        normalized_embeddings = embeddings / norms

        # Calculate the resultant vector R
        R = np.sum(normalized_embeddings, axis=0)

        # Calculate the mean resultant length R_bar (rho)
        N = len(embeddings)
        R_norm = np.linalg.norm(R)
        R_bar = R_norm / N

        # Trigger kill switch if semantic density exceeds threshold
        if R_bar > self.threshold:
            self.panic_mode = True
            return True

        return False

    def estimate_kappa(self, R_bar: float, d: int) -> float:
        """
        Numerically estimates the concentration parameter kappa of the von Mises-Fisher distribution.
        This solves A_d(kappa) = R_bar, where A_d(kappa) = I_{d/2}(kappa) / I_{d/2-1}(kappa).

        Args:
            R_bar (float): Mean resultant length.
            d (int): Dimensionality of the embeddings.

        Returns:
            float: Estimated kappa.
        """
        # Approximation for initial guess based on Banerjee et al. (2005)
        # kappa ~= (R_bar * d - R_bar**3) / (1 - R_bar**2)
        if R_bar >= 1.0:
            return float('inf')
        if R_bar <= 0.0:
            return 0.0

        kappa_guess = (R_bar * d - R_bar**3) / (1 - R_bar**2)

        # Function to find root for: f(kappa) = A_d(kappa) - R_bar = 0
        def f(k):
            # A_d(k) = I_{d/2}(k) / I_{d/2-1}(k)
            # scipy.special.ive is exponentially scaled modified Bessel function,
            # which avoids overflow for large kappa.
            # I_v(k) / I_{v-1}(k) == ive_v(k) / ive_{v-1}(k)
            num = scipy.special.ive(d / 2, k)
            den = scipy.special.ive(d / 2 - 1, k)
            if den == 0:
                # Fallback to standard iv if ive returns 0 (e.g. very small kappa)
                num = scipy.special.iv(d / 2, k)
                den = scipy.special.iv(d / 2 - 1, k)
                if den == 0:
                    return -R_bar # Should not happen for k>0, but safe fallback
            return (num / den) - R_bar

        # Optimization: use scipy.optimize.root_scalar
        try:
            from scipy.optimize import root_scalar
            # Newton or secant might be faster, but let's try a simple bracket first if possible,
            # or use the guess with a secant method.
            # Since A_d(k) is monotonically increasing from 0 to 1 as k goes 0 to inf:
            # For a given R_bar in (0, 1), there is a unique root.
            res = root_scalar(f, x0=kappa_guess, x1=kappa_guess * 1.1, method='secant')
            if res.converged:
                return res.root
        except Exception:
            pass

        # Fallback to simple Newton-Raphson if optimization fails or isn't available
        # derivative A_d'(k) = 1 - A_d(k)^2 - (d-1)/k * A_d(k)
        kappa = kappa_guess
        for _ in range(50): # Max iterations
            val = f(kappa)
            if abs(val) < 1e-6:
                break

            A_d = val + R_bar
            deriv = 1 - A_d**2 - ((d - 1) / kappa) * A_d

            if deriv == 0:
                break

            kappa = kappa - val / deriv

            # kappa must be > 0
            if kappa <= 0:
                kappa = 1e-4

        return kappa
