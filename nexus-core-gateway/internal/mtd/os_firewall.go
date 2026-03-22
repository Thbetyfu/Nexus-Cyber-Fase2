package mtd

import (
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// BlockIPAtOSLevel implements the 'Critical Architecture' kernel-level drop.
// It executes 'iptables -A INPUT -s <ip> -j DROP' on Linux systems.
// For Windows, it fails gracefully with a log entry (ISO-25010 Reliability).
func BlockIPAtOSLevel(ip string) {
	// Clean IP address input (remove port suffix if present)
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	if runtime.GOOS == "windows" {
		log.Printf("[OS-FIREWALL] Skipping iptables on Windows for IP: %s (OS Firewall bypassed)", ip)
		return
	}

	log.Printf("[OS-FIREWALL] KERNEL_DROP_INIT: Attempting to block %s via Netfilter...", ip)

	// Command: iptables -C INPUT -s <ip> -j DROP (Check if already exists)
	checkCmd := exec.Command("iptables", "-C", "INPUT", "-s", ip, "-j", "DROP")
	if err := checkCmd.Run(); err == nil {
		log.Printf("[OS-FIREWALL] %s is already dropped by kernel. Skipping duplicate.", ip)
		return
	}

	// Command: iptables -A INPUT -s <ip> -j DROP
	cmd := exec.Command("iptables", "-A", "INPUT", "-s", ip, "-j", "DROP")
	err := cmd.Run()
	if err != nil {
		// LOGGING ERROR (Could be missing NET_ADMIN or not on Linux)
		log.Printf("[OS-FIREWALL] KERNEL_DROP_ERROR: Failed to execute iptables for %s: %v", ip, err)
		return
	}

	log.Printf("[OS-FIREWALL] CRITICAL_DEFENSE: IP '%s' has been officially DROP-ed by Netfilter (Kernel Level).", ip)
}

// UnblockIPAtOSLevel (Bonus) - Remediate IP after threat duration
func UnblockIPAtOSLevel(ip string) {
	if runtime.GOOS == "windows" {
		return
	}

	cmd := exec.Command("iptables", "-D", "INPUT", "-s", ip, "-j", "DROP")
	if err := cmd.Run(); err != nil {
		log.Printf("[OS-FIREWALL] Failed to release IP %s: %v", ip, err)
	}
}
