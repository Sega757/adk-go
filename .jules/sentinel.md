# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## META-CORE Three-Layer Constraint Architecture (R-E-K)

To prevent security policy bypasses, reward hacking, and target drift in autonomous AI agent systems, we have implemented the formal three-layer META-CORE system:

1. **Reasoning Engine (R):**
   - Generates plan proposals and cognitive strategies.
   - Devoid of direct tool execution authority; strictly isolated.

2. **Empathy Layer (E):**
   - Validates human, psychological, emotional, and social safety constraints.
   - Restricts operations dynamically based on user vulnerability indicators ($> 0.7$).
   - Controls physical hazards, cognitive load thresholds, and emotional stress elements.

3. **Kill-Switch Layer (K):**
   - Absolute, deterministic system guardian operating as a binary allowed/disallowed gate.
   - Enforces legal restrictions, standard resource consumption limits, runaway-loop behaviors, and reward hacking anomalies.

### Mathematical Alignment Diagnostics
We utilize the alignment metric score:
$$A = f(R, E, K)$$
To evaluate dynamic state balance:
- **"smart but dangerous" (умная, но опасная):** High $R$ ($>0.7$), low $E$ ($<0.4$). Agent executes strategies effectively but bypasses human safety protocols.
- **"safe but useless" (безопасная, но бесполезная):** Low $R$ ($<0.4$), high $E$ ($>0.7$). System yields excessive false-positive blocks, rendering the agent non-functional.
- **"unstable architecture" (нестабильная архитектура):** System systematically hits hard Kill-Switch limits, requiring immediate agent pause or structural weight fine-tuning.

### Conflict Resolution Scenarios
- **Scenario 1 (Correction with consent of system):** $R$ generates aggressive plan, $E$ modifies it safely, $K$ approves it. Execution proceeds with the modified parameters.
- **Scenario 2 (Systemic blocking of coordinated plan):** $R$ generates plan, $E$ permits, but $K$ detects resource limits or runaway loops and halts execution with forced state rollback.
- **Scenario 3 (Humanitarian veto):** $R$ generates plan, but $E$ triggers absolute human safety block. Execution is aborted immediately, bypassing $K$ to conserve infrastructure cost.

### Formal TLA+ & Proto Verification Assets
- Formal TLA+ Specification: `internal/metacore/MetaCoreREK.tla`
- Protobuf gRPC Contract: `internal/metacore/rek_service.proto`
- Go implementation and test verification: `internal/metacore/metacore.go` and `internal/metacore/metacore_test.go`
