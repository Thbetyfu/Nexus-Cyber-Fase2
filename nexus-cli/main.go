package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ANSI Color Codes for Premium UI Feel
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
)

func printLogo() {
	fmt.Println(Cyan + Bold + `
    _   _ _______   ___    _  _____ 
   | \ | |  ___\ \ / / |  | |/ ____|
   |  \| | |__  \ V /| |  | | (___  
   | .   |  __|  > < | |  | |\___ \ 
   | |\  | |____/ . \| |__| |____) |
   |_| \_|______/_/ \_\____/|_____/ 
                                    
   === AUTONOMOUS CYBER DEFENSE ===
` + Reset)
}

func simulateBootSequence() {
	steps := []string{
		"Initializing PQC (Post-Quantum Cryptography) Module...",
		"Loading MTD (Moving Target Defense) Matrix...",
		"Establishing Secure Gateway...",
		"Synchronizing NEXUS-SOC-BRAIN AI...",
		"Activating Digital Hallucination Engine...",
	}

	fmt.Println(Cyan + "[SYSTEM] Initiating Core Boot Sequence..." + Reset)
	for _, step := range steps {
		time.Sleep(600 * time.Millisecond) // Simulated load time
		fmt.Printf("%s[*]%s %s %s[OK]%s\n", Blue, Reset, step, Green, Reset)
	}
	fmt.Println()
}

func runCompose(args ...string) error {
	cmd := exec.Command("docker-compose", args...)
	
	// Check if docker-compose.yml exists in the current directory
	if _, err := os.Stat("docker-compose.yml"); err == nil {
		// Found it in the current directory (Root)
		cmd.Dir = "."
	} else if _, err := os.Stat("../docker-compose.yml"); err == nil {
		// Found it one level up (if running from nexus-cli folder)
		cmd.Dir = ".."
	} else {
		return fmt.Errorf("docker-compose.yml not found. Please run this command from the Nexus-Cyber-Otonous root directory")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func statusCheck() {
	fmt.Println(Cyan + Bold + "\n=== NEXUS SYSTEM STATUS ===" + Reset)
	cmd := exec.Command("docker", "ps", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(Red + "Failed to get Docker status: " + err.Error() + Reset)
		return
	}
	
	output := string(out)
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		if strings.Contains(line, "nexus_") || strings.Contains(line, "target_portfolio") || strings.Contains(line, "NAMES") {
			if strings.Contains(line, "NAMES") {
				fmt.Println(Yellow + line + Reset)
			} else if strings.Contains(line, "Up") {
				fmt.Println(Green + line + Reset)
			} else {
				fmt.Println(Red + line + Reset)
			}
		}
	}
	fmt.Println()
}

func main() {
	if len(os.Args) < 2 {
		printLogo()
		fmt.Println("Usage: nexus [command]")
		fmt.Println("\nCommands:")
		fmt.Println("  up      - Boot the entire Nexus Cyber mesh autonomously")
		fmt.Println("  down    - Safely shut down all defense matrices")
		fmt.Println("  status  - Check operational status of all modules")
		fmt.Println("  logs    - Stream real-time defense logs")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "up":
		printLogo()
		simulateBootSequence()
		fmt.Println(Cyan + ">> Deploying Docker Containers..." + Reset)
		err := runCompose("up", "-d")
		if err != nil {
			fmt.Println(Red + "\n[ERROR] Deployment failed: " + err.Error() + Reset)
		} else {
			fmt.Println(Green + Bold + "\n[SUCCESS] Nexus Cyber is fully operational." + Reset)
			fmt.Println("Access Command Center at: " + Blue + "http://localhost:3000" + Reset)
		}
	case "down":
		fmt.Println(Yellow + "[SYSTEM] Initiating shutdown sequence..." + Reset)
		err := runCompose("down")
		if err != nil {
			fmt.Println(Red + "\n[ERROR] Shutdown failed: " + err.Error() + Reset)
		} else {
			fmt.Println(Green + "\n[SUCCESS] All defense matrices offline." + Reset)
		}
	case "status":
		statusCheck()
	case "logs":
		// Follow logs of the gateway by default
		fmt.Println(Cyan + ">> Streaming Nexus Gateway Logs..." + Reset)
		err := runCompose("logs", "-f", "gateway")
		if err != nil {
			fmt.Println(Red + "[ERROR] Failed to fetch logs." + Reset)
		}
	default:
		fmt.Printf(Red+"[ERROR] Unknown command: %s\n"+Reset, command)
	}
}
