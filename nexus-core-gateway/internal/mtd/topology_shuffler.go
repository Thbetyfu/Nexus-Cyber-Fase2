package mtd

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/nexus-cyber/nexus-core-gateway/internal/database"
	"github.com/nexus-cyber/nexus-core-gateway/internal/models"
)

// TargetBackend represents a proxying destination.
type TargetBackend struct {
	Host string
	Port int
}

// URL returns the full address string.
func (t TargetBackend) URL() string {
	return fmt.Sprintf("http://%s:%d", t.Host, t.Port)
}

// TopologyShuffler implements @skill-mtd Dynamic Configuration Randomization.
// It rotates the backend port periodically using a CSPRNG to ensure unpredictability.
type TopologyShuffler struct {
	mu           sync.RWMutex
	current      TargetBackend
	baseHost     string
	portPool     []int // Allowed ports to rotate between
	intervalSecs int
	stopCh       chan struct{}
	lastShuffle  time.Time
	onShuffle    func(newTarget TargetBackend) // callback hook for graceful handoff
}

// NewTopologyShuffler creates and starts a new MTD topology shuffler.
// portPool: list of valid backend ports to cycle between.
// intervalSecs: how often to rotate (e.g., 60 seconds).
func NewTopologyShuffler(baseHost string, portPool []int, intervalSecs int, onShuffle func(TargetBackend)) *TopologyShuffler {
	if len(portPool) == 0 {
		panic("mtd: portPool must not be empty")
	}

	// Select initial port using CSPRNG
	initialPort := portPool[csrng(len(portPool))]

	ts := &TopologyShuffler{
		baseHost:     baseHost,
		portPool:     portPool,
		intervalSecs: intervalSecs,
		current:      TargetBackend{Host: baseHost, Port: initialPort},
		stopCh:       make(chan struct{}),
		lastShuffle:  time.Now(),
		onShuffle:    onShuffle,
	}

	return ts
}

// Start begins the periodic rotation goroutine.
func (ts *TopologyShuffler) Start() {
	go ts.rotationLoop()
	log.Printf("[MTD] TopologyShuffler ACTIVE — Rotating every %ds across %d port targets", ts.intervalSecs, len(ts.portPool))
}

// Stop halts the rotation loop gracefully.
func (ts *TopologyShuffler) Stop() {
	close(ts.stopCh)
}

// GetCurrent returns the current active backend target (thread-safe).
func (ts *TopologyShuffler) GetCurrent() TargetBackend {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.current
}

// GetStatus returns the current port and seconds until next shuffle.
func (ts *TopologyShuffler) GetStatus() (int, int) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	elapsed := time.Since(ts.lastShuffle).Seconds()
	remain := ts.intervalSecs - int(elapsed)
	if remain < 0 {
		remain = 0
	}
	return ts.current.Port, remain
}

// rotationLoop is the internal goroutine that triggers port rotation.
func (ts *TopologyShuffler) rotationLoop() {
	ticker := time.NewTicker(time.Duration(ts.intervalSecs) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ts.mu.Lock()
			oldPort := ts.current.Port
			// CSPRNG selection — exclude current port to guarantee rotation
			newPort := ts.selectNewPort(oldPort)
			ts.current = TargetBackend{Host: ts.baseHost, Port: newPort}
			ts.lastShuffle = time.Now()
			newTarget := ts.current
			ts.mu.Unlock()

			log.Printf("[MTD-SHUFFLE] Backend rotated: :%d -> :%d | Time: %s",
				oldPort, newPort, time.Now().Format(time.RFC3339))

			// Graceful Handoff: notify proxy core without dropping in-flight connections
			if ts.onShuffle != nil {
				go ts.onShuffle(newTarget) // async to not block rotation loop
			}

			// 3. PERSISTENCE: Log audit trail to Database (ISO 27001)
			if database.DB != nil {
				go func(o, n int) {
					audit := models.MTDAuditTrail{
						OldPort:       o,
						NewPort:       n,
						TriggerReason: "SCHEDULED_ROTATION",
						Status:        "SUCCESS",
					}
					database.DB.Create(&audit)
				}(oldPort, newPort)
			}

		case <-ts.stopCh:
			log.Printf("[MTD] TopologyShuffler STOPPED.")
			return
		}
	}
}

// selectNewPort uses CSPRNG to pick a port from the pool, excluding the current one.
func (ts *TopologyShuffler) selectNewPort(currentPort int) int {
	// Build candidate pool excluding current port
	candidates := make([]int, 0, len(ts.portPool)-1)
	for _, p := range ts.portPool {
		if p != currentPort {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return currentPort // only one option
	}
	return candidates[csrng(len(candidates))]
}

// ManualShuffle triggers an immediate, unpredicted topology rotation.
// Part of the @skill-self-repair "Rescue Protocol".
func (ts *TopologyShuffler) ManualShuffle() {
	ts.mu.Lock()
	oldPort := ts.current.Port
	newPort := ts.selectNewPort(oldPort)
	ts.current = TargetBackend{Host: ts.baseHost, Port: newPort}
	ts.lastShuffle = time.Now()
	newTarget := ts.current
	ts.mu.Unlock()

	log.Printf("[MTD-PANIC] EMERGENCY SHUFFLE TRIGGERED: :%d -> :%d", oldPort, newPort)

	if ts.onShuffle != nil {
		go ts.onShuffle(newTarget)
	}

	// PERSISTENCE: Log audit trail to Database (ISO 27001)
	if database.DB != nil {
		go func(o, n int) {
			audit := models.MTDAuditTrail{
				OldPort:       o,
				NewPort:       n,
				TriggerReason: "EMERGENCY_MANUAL_SHUFFLE",
				Status:        "SUCCESS",
			}
			database.DB.Create(&audit)
		}(oldPort, newPort)
	}
}

// csrng returns a cryptographically secure random int in [0, n).
func csrng(n int) int {
	max := big.NewInt(int64(n))
	val, err := rand.Int(rand.Reader, max)
	if err != nil {
		// Fallback: should never happen in practice
		panic(fmt.Sprintf("mtd: CSPRNG failure: %v", err))
	}
	return int(val.Int64())
}
