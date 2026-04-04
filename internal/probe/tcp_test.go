package probe

import (
	"bufio"
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/AnubisWatch/anubiswatch/internal/core"
)

func TestTCPChecker_Judge_Basic(t *testing.T) {
	// Create TCP listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Start goroutine to accept connections
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	checker := NewTCPChecker()

	soul := &core.Soul{
		ID:      "test-tcp",
		Name:    "Test TCP",
		Type:    core.CheckTCP,
		Target:  listener.Addr().String(),
		Enabled: true,
		Weight:  core.Duration{Duration: 60 * time.Second},
		TCP:     &core.TCPConfig{},
	}

	ctx := context.Background()
	judgment, err := checker.Judge(ctx, soul)

	if err != nil {
		t.Fatalf("Judge failed: %v", err)
	}

	if judgment.Status != core.SoulAlive {
		t.Errorf("Expected status Alive, got %s", judgment.Status)
	}
}

func TestTCPChecker_Judge_ConnectionRefused(t *testing.T) {
	checker := NewTCPChecker()

	soul := &core.Soul{
		ID:     "test-tcp",
		Name:   "Test TCP",
		Type:   core.CheckTCP,
		Target: "127.0.0.1:1", // Invalid port
		TCP:    &core.TCPConfig{},
	}

	ctx := context.Background()
	judgment, _ := checker.Judge(ctx, soul)

	if judgment.Status != core.SoulDead {
		t.Errorf("Expected status Dead, got %s", judgment.Status)
	}
}

func TestTCPChecker_Judge_BannerMatch(t *testing.T) {
	// Create TCP listener that sends banner
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				c.Write([]byte("SSH-2.0-OpenSSH_8.0\r\n"))
				io.Copy(io.Discard, c)
			}(conn)
		}
	}()

	checker := NewTCPChecker()

	soul := &core.Soul{
		ID:     "test-tcp",
		Name:   "Test TCP",
		Type:   core.CheckTCP,
		Target: listener.Addr().String(),
		TCP: &core.TCPConfig{
			BannerMatch: "OpenSSH",
		},
	}

	ctx := context.Background()
	judgment, _ := checker.Judge(ctx, soul)

	if judgment.Status != core.SoulAlive {
		t.Errorf("Expected status Alive, got %s", judgment.Status)
	}
}

func TestTCPChecker_Judge_BannerMismatch(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				c.Write([]byte("SSH-2.0-OpenSSH_8.0\r\n"))
				io.Copy(io.Discard, c)
			}(conn)
		}
	}()

	checker := NewTCPChecker()

	soul := &core.Soul{
		ID:     "test-tcp",
		Name:   "Test TCP",
		Type:   core.CheckTCP,
		Target: listener.Addr().String(),
		TCP: &core.TCPConfig{
			BannerMatch: "Postfix", // Wrong banner
		},
	}

	ctx := context.Background()
	judgment, _ := checker.Judge(ctx, soul)

	if judgment.Status != core.SoulDead {
		t.Errorf("Expected status Dead, got %s", judgment.Status)
	}
}

func TestTCPChecker_Judge_ExpectRegex(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				c.Write([]byte("220 mail.example.com ESMTP\r\n"))
				io.Copy(io.Discard, c)
			}(conn)
		}
	}()

	checker := NewTCPChecker()

	soul := &core.Soul{
		ID:     "test-tcp",
		Name:   "Test TCP",
		Type:   core.CheckTCP,
		Target: listener.Addr().String(),
		TCP: &core.TCPConfig{
			ExpectRegex: "^220.*ESMTP",
		},
	}

	ctx := context.Background()
	judgment, _ := checker.Judge(ctx, soul)

	if judgment.Status != core.SoulAlive {
		t.Errorf("Expected status Alive, got %s", judgment.Status)
	}
}

func TestTCPChecker_Judge_ExpectRegexMismatch(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				c.Write([]byte("220 mail.example.com ESMTP\r\n"))
				io.Copy(io.Discard, c)
			}(conn)
		}
	}()

	checker := NewTCPChecker()

	soul := &core.Soul{
		ID:     "test-tcp",
		Name:   "Test TCP",
		Type:   core.CheckTCP,
		Target: listener.Addr().String(),
		TCP: &core.TCPConfig{
			ExpectRegex: "^550.*Relay denied", // Wrong pattern
		},
	}

	ctx := context.Background()
	judgment, _ := checker.Judge(ctx, soul)

	if judgment.Status != core.SoulDead {
		t.Errorf("Expected status Dead, got %s", judgment.Status)
	}
}

func TestTCPChecker_Judge_SendExpect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				reader := bufio.NewReader(c)
				line, _ := reader.ReadString('\n')
				if line == "PING\n" {
					c.Write([]byte("PONG\n"))
				}
			}(conn)
		}
	}()

	checker := NewTCPChecker()

	soul := &core.Soul{
		ID:     "test-tcp",
		Name:   "Test TCP",
		Type:   core.CheckTCP,
		Target: listener.Addr().String(),
		TCP: &core.TCPConfig{
			Send:        "PING\n",
			ExpectRegex: "PONG",
		},
	}

	ctx := context.Background()
	judgment, _ := checker.Judge(ctx, soul)

	if judgment.Status != core.SoulAlive {
		t.Errorf("Expected status Alive, got %s", judgment.Status)
	}
}

func TestTCPChecker_Validate_MissingTarget(t *testing.T) {
	checker := NewTCPChecker()

	soul := &core.Soul{
		ID:   "test-tcp",
		Name: "Test TCP",
		Type: core.CheckTCP,
	}

	err := checker.Validate(soul)
	if err == nil {
		t.Error("Expected validation error for missing target")
	}
}

func TestTCPChecker_Validate_InvalidFormat(t *testing.T) {
	checker := NewTCPChecker()

	soul := &core.Soul{
		ID:     "test-tcp",
		Name:   "Test TCP",
		Type:   core.CheckTCP,
		Target: "invalid-no-port",
	}

	err := checker.Validate(soul)
	if err == nil {
		t.Error("Expected validation error for invalid format")
	}
}

func TestTCPChecker_Validate_Valid(t *testing.T) {
	checker := NewTCPChecker()

	soul := &core.Soul{
		ID:     "test-tcp",
		Name:   "Test TCP",
		Type:   core.CheckTCP,
		Target: "localhost:443",
	}

	err := checker.Validate(soul)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestUDPChecker_Validate_MissingTarget(t *testing.T) {
	checker := NewUDPChecker()

	soul := &core.Soul{
		ID:   "test-udp",
		Name: "Test UDP",
		Type: core.CheckUDP,
	}

	err := checker.Validate(soul)
	if err == nil {
		t.Error("Expected validation error for missing target")
	}
}

func TestUDPChecker_Validate_InvalidFormat(t *testing.T) {
	checker := NewUDPChecker()

	soul := &core.Soul{
		ID:     "test-udp",
		Name:   "Test UDP",
		Type:   core.CheckUDP,
		Target: "invalid-no-port",
	}

	err := checker.Validate(soul)
	if err == nil {
		t.Error("Expected validation error for invalid format")
	}
}

func TestUDPChecker_Validate_Valid(t *testing.T) {
	checker := NewUDPChecker()

	soul := &core.Soul{
		ID:     "test-udp",
		Name:   "Test UDP",
		Type:   core.CheckUDP,
		Target: "localhost:53",
	}

	err := checker.Validate(soul)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestUDPChecker_Judge_InvalidHexPayload(t *testing.T) {
	checker := NewUDPChecker()

	soul := &core.Soul{
		ID:     "test-udp",
		Name:   "Test UDP",
		Type:   core.CheckUDP,
		Target: "127.0.0.1:53",
		UDP: &core.UDPConfig{
			SendHex: "invalid hex !!!",
		},
	}

	ctx := context.Background()
	judgment, _ := checker.Judge(ctx, soul)

	if judgment.Status != core.SoulDead {
		t.Errorf("Expected status Dead, got %s", judgment.Status)
	}
}

func TestUDPChecker_Judge_NoResponse(t *testing.T) {
	// Bind to a port but don't respond
	listener, err := net.Listen("udp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("UDP listener not supported on this platform: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	checker := NewUDPChecker()

	soul := &core.Soul{
		ID:      "test-udp",
		Name:    "Test UDP",
		Type:    core.CheckUDP,
		Target:  addr,
		Timeout: core.Duration{Duration: 100 * time.Millisecond},
		UDP: &core.UDPConfig{
			SendHex: "0001",
		},
	}

	ctx := context.Background()
	judgment, _ := checker.Judge(ctx, soul)

	// UDP is connectionless, so no response means dead
	if judgment.Status != core.SoulDead {
		t.Errorf("Expected status Dead, got %s", judgment.Status)
	}
}
