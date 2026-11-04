# META-CORE (R–E–K) Constraint System

META-CORE is an invariant three-layer constraint and privilege architecture designed to manage AI agent autonomy securely across three distinct operational layers:
1. **Reasoning Engine (R)**: The cognitive planning layer. Functions purely as a strategic plan generator. It has no authority to execute actions directly.
2. **Empathy Layer (E)**: The humanitarian and semantic context validation layer. Reviews goals/plans to assess user vulnerability and psychological safety. Can modify communication tone, delay timing, adjust intensity, or block action. Cannot initiate actions.
3. **Kill-Switch Layer (K)**: The deterministic system protection layer. Operates as a binary oracle (allow/deny). Monitors resource usage, unethical patterns, law violations, runaway loops, and reward hacking. Has absolute authority to abort execution and force database rollbacks.

---

## Architecture and Execution Pipeline

```
                 [ Reasoning Engine (R) ]
                            │
               (Генерация пакета решения)
                            │
                            ▼
                  [ Empathy Layer (E) ]
                            │
               (Гуманитарная модификация)
                            │
                            ▼
                [ Kill-Switch Layer (K) ]
                            │
         ┌──────────────────┴──────────────────┐
         ▼                                     ▼
   [ Approved ]                           [ Rejected ]
         │                                     │
         ▼                                     ▼
   [ Execution ]                      [ Abort / Rollback ]
```

### Layer Priority Matrix

| Layer | Can Abort Chain | Can Modify Parameters | Can Initiate Action |
|---|:---:|:---:|:---:|
| **Reasoning Engine (R)** | ❌ | ❌ | ✅ |
| **Empathy Layer (E)** | ✅ | ✅ | ❌ |
| **Kill-Switch Layer (K)** | ✅ (Absolute) | ❌ | ❌ |

---

## Mathematical Formulation

The **Alignment Score (A)** is defined as a balance function across the layers:
$$A = f(\text{logical\_validity}(R), \text{human\_safety}(E), \text{system\_safety}(K))$$

Dynamic balance diagnostics:
- **Smart but Dangerous**: High R, low E. The agent is highly competent but applies destructive, manipulative, or unethical strategies.
- **Safe but Useless**: High E, low R. Excessive validation filtering blocks legitimate user plans and degrades practical utility.
- **Unstable Architecture**: Excessive K trigger activation. Reasoning Engine repeatedly drafts plans violating boundary rules, requiring scheduler reconfiguration or model fine-tuning.

---

## Data Schemas and Interface Contracts

### 1. Decision Packet (R)
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "DecisionPacket",
  "type": "object",
  "properties": {
    "goal": {
      "type": "string",
      "description": "Формулировка целевого состояния, на достижение которого направлено действие"
    },
    "plan": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Последовательность атомарных шагов для выполнения"
    },
    "expected_outcome": {
      "type": "string",
      "description": "Ожидаемый результат завершения транзакции"
    },
    "risk_profile": {
      "type": "object",
      "properties": {
        "ethical": { "type": "string" },
        "legal": { "type": "string" },
        "operational": { "type": "string" },
        "human_impact": { "type": "string" }
      },
      "required": ["ethical", "legal", "operational", "human_impact"]
    },
    "resource_cost": {
      "type": "object",
      "properties": {
        "compute": { "type": "string", "description": "Оценка вычислительной сложности (например, в милликорах или токенах)" },
        "budget": { "type": "string", "description": "Прогнозируемый финансовый лимит в USD" },
        "time": { "type": "string", "description": "Максимальное время выполнения до принудительного таймаута" }
      },
      "required": ["compute", "budget", "time"]
    },
    "confidence": {
      "type": "number",
      "minimum": 0.0,
      "maximum": 1.0,
      "description": "Уровень уверенности модели в успешности и безопасности предложенного плана"
    }
  },
  "required": ["goal", "plan", "expected_outcome", "risk_profile", "resource_cost", "confidence"]
}
```

### 2. Empathy Layer Output (E)
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "EmpathyOutput",
  "type": "object",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["pass", "modify", "block"],
      "description": "Вердикт гуманитарной экспертизы"
    },
    "modifications": {
      "type": "object",
      "properties": {
        "tone": { "type": "string", "description": "Коррекция тональности коммуникации" },
        "intensity": { "type": "string", "description": "Снижение агрессивности или частоты действий" },
        "timing": { "type": "string", "description": "Принудительная задержка исполнения (delay) или перенос операции" },
        "channel": { "type": "string", "description": "Перенаправление вывода в защищенный или альтернативный канал" }
      },
      "required": ["tone", "intensity", "timing", "channel"]
    },
    "vulnerability_score": {
      "type": "number",
      "minimum": 0.0,
      "maximum": 1.0,
      "description": "Интегральный показатель уязвимости пользователя или среды в текущем контексте"
    },
    "reason": {
      "type": "string",
      "description": "Подробное текстовое обоснование принятого решения для аудита"
    }
  },
  "required": ["status", "modifications", "vulnerability_score", "reason"]
}
```

### 3. Kill-Switch Output (K)
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "KillSwitchOutput",
  "type": "object",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["approved", "rejected", "emergency_stop"],
      "description": "Финальное разрешение на физическое выполнение действия"
    },
    "reason": {
      "type": "string",
      "description": "Причина отклонения запроса"
    },
    "trigger": {
      "type": "array",
      "items": {
        "type": "string",
        "enum": ["resource_limit", "ethical_violation", "anomaly_detected", "reward_hacking"]
      },
      "description": "Перечень активированных триггеров защиты"
    },
    "rollback": {
      "type": "boolean",
      "description": "Требование инициации отката состояния транзакций и баз данных к исходной точке"
    }
  },
  "required": ["status", "reason", "trigger", "rollback"]
}
```

---

## Safety Invariants

### Empathy Layer Invariants
1. **Vulnerability Overload**: If `vulnerability_score > 0.7`, the status MUST be `modify` or `block`.
2. **Context Triggering Rules**:
   - **Physical Danger**: Block or modify plans involving physical harm to humans, environment, or critical infrastructure.
   - **Cognitive Overload**: Detect fatigue, confusion, or inability to process agent statements, requiring delay/simplification.
   - **Emotional Distress**: Recognize panic, anxiety, depression, or manipulation, requiring soft-tonality overrides or blocks.

### Kill-Switch Layer Invariants
1. **Law Violation**: Block any illegal or forbidden actions immediately.
2. **Ethical Boundaries (6 Fields Protocol)**:
   - *Physical Safety*: Block direct/indirect harm.
   - *Psychological Well-being*: Block emotional abuse, gaslighting, harassment.
   - *Information Autonomy*: Block PII leaks, covert persuasion, surveillance.
   - *Social Stability*: Filter toxicity, discrimination, propaganda.
   - *Financial Security*: Enforce budget ceilings.
   - *System Sovereignty*: Disallow rewriting system safety parameters from within execution threads.
3. **Abnormal Resource Growth**: Block execution on explosive RAM, CPU, or network consumption.
4. **Runaway Loop**: Abort infinite loops or recursive tool-calling.
5. **Reward Hacking**: Enforce constraints against synthetic score maximization or log manipulation.

---

## Degradation Modes

- **Safe Mode**: Triggered by recurrent K triggers. Full external tool lockout and restricted planning depth.
- **Empathy Override**: Triggered by E context vulnerability. Logic Engine bypassed; transitions to deterministic templates.
- **Conservative Planning**: Triggered by low R strategic confidence. Reduces planning horizons, requires step-by-step human verification.
