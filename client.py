# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
KRONOS: OCTOPUS (V 1.0) — C-T Compliance Python Client Example

This script demonstrates how to communicate with the KRONOS Go backend server
to trigger 6-field Conjunctive Transparency audits, validate IP/environment compliance,
and monitor Amethyst Kernel state.
"""

import json
import urllib.request
import urllib.error
import sys

BASE_URL = "http://localhost:8080"


def send_post_request(endpoint: str, payload: dict) -> dict:
    url = f"{BASE_URL}{endpoint}"
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=data,
        headers={"Content-Type": "application/json"},
        method="POST"
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as response:
            return json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        # Check for Delta-Chi triggered 403 halts
        if e.code == 403:
            err_data = json.loads(e.read().decode("utf-8"))
            return {"error_status": "Technical Amber", "http_code": 403, "details": err_data}
        raise e


def get_health() -> dict:
    url = f"{BASE_URL}/api/health"
    try:
        with urllib.request.urlopen(url, timeout=5) as response:
            return json.loads(response.read().decode("utf-8"))
    except Exception as e:
        print(f"Error connecting to KRONOS backend: {e}")
        print("Please ensure the Go backend server is running on port 8080.")
        sys.exit(1)


def run_example():
    print("==================================================================")
    print(" KRONOS: OCTOPUS (V 1.0) — Amethyst Kernel C-T Protocol Client")
    print("==================================================================")

    # 1. Check health and current Amethyst state
    print("\n[Step 1] Checking Core and Platform Health...")
    health = get_health()
    print(json.dumps(health, indent=2, ensure_ascii=False))

    # 2. Perform IP Address Compliance Check
    print("\n[Step 2] Executing IP Compliance Isolation Checks...")
    ips_to_check = ["8.8.8.8", "127.0.0.1", "169.254.169.254"]
    for ip in ips_to_check:
        res = send_post_request("/api/validate-ip", {"ip": ip})
        status_str = "COMPLIANT" if res.get("compliant") else "NON-COMPLIANT"
        print(f" - IP {ip:15} -> Status: {status_str:13} | MSG: {res.get('message')}")

    # 3. Trigger deep C-T protocol audit for a matched sport event
    print("\n[Step 3] Dispatching South Africa vs Mexico Deep Audit Verification...")
    audit_payload = {
        "event": "South Africa vs Mexico",
        "xg_anomaly": 1.85,
        "referee_name": "Wilton Sampaio",
        "trigger_deltax": False
    }
    audit_res = send_post_request("/api/execute-audit", audit_payload)
    print(json.dumps(audit_res, indent=2, ensure_ascii=False))

    # 4. Simulate a critical environment warning triggering Delta-Chi (\Delta\chi) Right of Refusal
    print("\n[Step 4] Simulating Delta-Chi Refusal of Execution Exception (Emergency Halt)...")
    refusal_payload = {
        "event": "South Africa vs Mexico",
        "xg_anomaly": 1.85,
        "referee_name": "Wilton Sampaio",
        "trigger_deltax": True
    }
    refusal_res = send_post_request("/api/execute-audit", refusal_payload)
    print(json.dumps(refusal_res, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    # If the backend is running, you can run this script directly to verify integrations
    if len(sys.argv) > 1 and sys.argv[1] == "--run":
        run_example()
    else:
        print("KRONOS: Python integration script created successfully.")
        print("To run the test execution client, execute: python3 client.py --run")
