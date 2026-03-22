import requests
import time
import sys
import os

# Configuration
GATEWAY_URL = "http://localhost:8080"
TELEMETRY_API = f"{GATEWAY_URL}/api/telemetry"
TEST_DOMAIN = "ojk.localhost"

def print_result(feature, success, message):
    status = "SUCCESS" if success else "FAILED"
    color = "\033[92m" if success else "\033[91m"
    reset = "\033[0m"
    print(f"[{color}{status}{reset}] {feature}: {message}")

def audit_system():
    print("--- 🚀 NEXUS CYBER: SYSTEM INTEGRITY AUDIT ---")
    
    # 1. Connectivity Check
    try:
        res = requests.get(TELEMETRY_API, timeout=5)
        res.raise_for_status()
        data = res.json()
        print_result("Connectivity", True, "Gateway Telemetry API is reachable.")
    except Exception as e:
        print_result("Connectivity", False, f"Could not reach Gateway API: {e}")
        return

    # 2. Stats Check
    initial_allowed = data.get("stats", {}).get("allowed", 0)
    print_result("Baseline Stats", True, f"Initial Allowed traffic: {initial_allowed}")

    # 3. End-to-End Traffic Simulation
    print(f"[INFO] Simulating test request to http://{TEST_DOMAIN} via Gateway...")
    try:
        # We use Host header to simulate multi-tenant routing
        headers = {"Host": TEST_DOMAIN}
        # Simulate a small attack to check MTD/Honeypot as well
        simulation_res = requests.get(GATEWAY_URL, headers=headers, timeout=5)
        print_result("Routing", simulation_res.status_code == 200, f"Gateway routed {TEST_DOMAIN} with status {simulation_res.status_code}")
    except Exception as e:
        print_result("Routing", False, f"Request simulation failed: {e}")

    # 4. Verification Check
    print("[INFO] Waiting for telemetry sync...")
    time.sleep(2)
    
    try:
        res = requests.get(TELEMETRY_API, timeout=5)
        new_data = res.json()
        new_allowed = new_data.get("stats", {}).get("allowed", 0)
        
        increments = new_allowed > initial_allowed
        print_result("Telemetry Tracking", increments, f"Allowed counter incremented ({initial_allowed} -> {new_allowed})")
        
        if not increments:
            print("[CRITICAL] Statistics are NOT incrementing even after traffic simulation.")
    except Exception as e:
        print_result("Telemetry Tracking", False, f"Verification poll failed: {e}")

    # 5. Dashboard Hook Check
    recents = data.get("recent_logs", [])
    has_recent = len(recents) > 0
    print_result("Data Hub", has_recent, f"Found {len(recents)} logs in the telemetry buffer.")

    print("--- AUDIT COMPLETE ---")
    if increments and has_recent:
        print("\033[92mSYSTEM IS FUNCTIONAL. Metrics are flowing to the dashboard.\033[0m")
    else:
        print("\033[91mSYSTEM INTEGRITY COMPROMISED. Check Gateway logs for Routing Errors.\033[0m")

if __name__ == "__main__":
    audit_system()
