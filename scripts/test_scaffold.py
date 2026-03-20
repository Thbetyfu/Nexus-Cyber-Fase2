import os
import sys

def check_structure():
    print("🛡️ Nexus Cyber - Scaffold Verification Script")
    print("-" * 40)
    
    required_dirs = [
        "nexus-core-gateway/cmd/gateway",
        "nexus-core-gateway/internal/ai",
        "nexus-core-gateway/internal/mtd",
        "nexus-core-gateway/internal/crypto",
        "nexus-core-gateway/internal/proxy",
        "nexus-core-gateway/internal/repair",
        "nexus-admin-dashboard/app",
        ".agents/skills/dual-brain",
        ".agents/skills/qa-iso-auditor",
        "scripts",
        "docs"
    ]
    
    missing = []
    base_path = "d:/0. Kerjaan/Nexus-Cyber"
    
    for d in required_dirs:
        full_path = os.path.join(base_path, d)
        if os.path.exists(full_path):
            print(f"✅ {d:.<40} FOUND")
        else:
            print(f"❌ {d:.<40} MISSING")
            missing.append(d)
            
    print("-" * 40)
    if not missing:
        print("🎉 SCAFFOLD INTEGRITY: 100% - ALL SYSTEMS GO")
        return True
    else:
        print(f"⚠️ MISSING {len(missing)} COMPONENTS. RE-RUN SETUP.")
        return False

if __name__ == "__main__":
    success = check_structure()
    if not success:
        sys.exit(1)
