import json
import time
import os
from jsonschema import validate, ValidationError

# Жесткая привязка к схеме META-CORE
SCHEMA_PATH = os.path.join(os.path.dirname(__file__), "schemas/event_bus_payload.v1.json")
LOG_FILE = "/tmp/metacore_telemetry.jsonl"

def load_schema():
    with open(SCHEMA_PATH, 'r') as f:
        return json.load(f)

def tail_file(filepath):
    while not os.path.exists(filepath):
        time.sleep(0.5)

    with open(filepath, 'r') as f:
        f.seek(0, 2)  # Прыжок в конец файла
        while True:
            line = f.readline()
            if not line:
                time.sleep(0.1)
                continue
            yield line.strip()

def run_transponder():
    schema = load_schema()
    print("[FLAT.SYNTH] Transponder active. Listening for R-E-K telemetry...")

    for line in tail_file(LOG_FILE):
        if not line: continue
        try:
            event = json.loads(line)
            validate(instance=event, schema=schema)
            # Успешная валидация -> передача в observer или аналитику
            print(f"[OK] Event: {event['event_id']} | Type: {event['type']}")
        except ValidationError as e:
            print(f"[ERROR] Schema violation in {event.get('event_id', 'UNKNOWN')}: {e.message}")
        except json.JSONDecodeError:
            print("[ERROR] Corrupted JSON payload detected")

if __name__ == "__main__":
    run_transponder()
