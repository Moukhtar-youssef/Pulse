package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Moukhtar-youssef/Pulse/internal/network"
	"github.com/Moukhtar-youssef/Pulse/internal/protocol"
	"github.com/Moukhtar-youssef/Pulse/internal/utils"
)

const (
	serverHost        = "127.0.0.1"
	serverPort        = 8080
	metricsInterval   = 5 * time.Second
	initialRetryDelay = 2 * time.Second
	maxRetryDelay     = 60 * time.Second
	retryMultiplier   = 2
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[PULSE] ")

	// Context cancelled on SIGTERM / SIGINT (systemd sends SIGTERM on stop)
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("PULSE daemon starting...")

	runWithRetry(ctx)

	log.Println("PULSE daemon stopped.")
}

// runWithRetry keeps reconnecting with exponential backoff until ctx is done.
func runWithRetry(ctx context.Context) {
	delay := initialRetryDelay

	for {
		// Check if we should stop before trying
		select {
		case <-ctx.Done():
			return
		default:
		}

		log.Printf("Connecting to %s:%d...", serverHost, serverPort)

		conn := network.NewTCPConnection()
		err := conn.Connect(serverHost, serverPort)
		if err != nil {
			log.Printf("Connection failed: %v — retrying in %s", err, delay)
			waitOrExit(ctx, delay)
			delay = nextDelay(delay)
			continue
		}

		log.Println("Connected.")
		delay = initialRetryDelay // reset backoff on success

		// Run the main agent loop — returns when connection drops
		if err := runAgent(ctx, conn); err != nil {
			log.Printf("Agent error: %v", err)
		}

		conn.Close()

		// If context is done (shutdown), don't retry
		select {
		case <-ctx.Done():
			return
		default:
			log.Printf("Disconnected — reconnecting in %s", delay)
			waitOrExit(ctx, delay)
			delay = nextDelay(delay)
		}
	}
}

// runAgent sends the handshake then streams metrics until ctx or error.
func runAgent(ctx context.Context, conn *network.TCPConnection) error {
	hostname, _ := os.Hostname()
	hello := fmt.Sprintf("hostname=%s os=%s", hostname, utils.CurrentOS())

	if err := conn.SendPacket(protocol.TypeHello, hello); err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}
	log.Printf("Handshake sent: %s", hello)

	ticker := time.NewTicker(metricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil // clean shutdown

		case <-ticker.C:
			metrics := collectMetrics()
			if err := conn.SendPacket(protocol.TypeMetrics, metrics); err != nil {
				return fmt.Errorf("metrics send: %w", err)
			}
			if err := conn.SendPacket(protocol.TypeHeartbeat, "alive"); err != nil {
				return fmt.Errorf("heartbeat send: %w", err)
			}
			log.Printf("Sent: %s", metrics)
		}
	}
}

func collectMetrics() string {
	return "cpu=21 mem=44 disk=10"
}

func waitOrExit(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

func nextDelay(d time.Duration) time.Duration {
	d *= retryMultiplier
	if d > maxRetryDelay {
		d = maxRetryDelay
	}
	return d
}
