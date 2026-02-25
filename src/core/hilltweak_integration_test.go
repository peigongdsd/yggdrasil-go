package core_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/yggdrasil-network/yggdrasil-go/src/admin"
	"github.com/yggdrasil-network/yggdrasil-go/src/config"
)

type testNode struct {
	name         string
	cfgPath      string
	adminSock    string
	hillTweakMs  int64
	publicKeyHex string
	cmd          *exec.Cmd
	logs         *bytes.Buffer
}

func TestHillTweakHandshakeSum(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if runtime.GOOS == "windows" {
		t.Skip("unix admin sockets required for this test")
	}

	repoRoot := filepath.Clean(filepath.Join(mustGetwd(t), "..", ".."))
	bin := filepath.Join(t.TempDir(), "yggdrasil")
	build := exec.Command("go", "build", "-o", bin, "./cmd/yggdrasil")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build yggdrasil: %v\n%s", err, string(out))
	}

	tempDir := t.TempDir()
	ports := reservePorts(t, 3)
	nodes := make([]*testNode, 3)
	hills := []int64{100, 250, 400}
	for i := range nodes {
		adminSock := filepath.Join(tempDir, fmt.Sprintf("admin-%d.sock", i))
		cfg := config.GenerateConfig()
		cfg.IfName = "none"
		cfg.AdminListen = "unix://" + adminSock
		cfg.Listen = []string{fmt.Sprintf("tcp://127.0.0.1:%d", ports[i])}
		cfg.Peers = nil
		cfg.InterfacePeers = map[string][]string{}
		cfg.MulticastInterfaces = []config.MulticastInterfaceConfig{}
		cfg.AllowedPublicKeys = []string{}
		cfg.HillTweakMs = hills[i]

		cfgPath := filepath.Join(tempDir, fmt.Sprintf("node-%d.json", i))
		writeConfig(t, cfgPath, cfg)

		pub := ed25519.PrivateKey(cfg.PrivateKey).Public().(ed25519.PublicKey)
		nodes[i] = &testNode{
			name:         fmt.Sprintf("node-%d", i),
			cfgPath:      cfgPath,
			adminSock:    adminSock,
			hillTweakMs:  hills[i],
			publicKeyHex: hex.EncodeToString(pub),
			logs:         &bytes.Buffer{},
		}
	}

	// Ring topology: 0 -> 1 -> 2 -> 0
	nodes[0].cfgPath = setPeers(t, nodes[0].cfgPath, []string{fmt.Sprintf("tcp://127.0.0.1:%d", ports[1])})
	nodes[1].cfgPath = setPeers(t, nodes[1].cfgPath, []string{fmt.Sprintf("tcp://127.0.0.1:%d", ports[2])})
	nodes[2].cfgPath = setPeers(t, nodes[2].cfgPath, []string{fmt.Sprintf("tcp://127.0.0.1:%d", ports[0])})

	for _, node := range nodes {
		startNode(t, repoRoot, bin, node)
	}
	t.Cleanup(func() {
		for _, node := range nodes {
			stopNode(t, node)
		}
	})

	for _, node := range nodes {
		waitForAdmin(t, node.adminSock, 10*time.Second)
	}

	peersByKey := map[string]int64{}
	for _, node := range nodes {
		peersByKey[node.publicKeyHex] = node.hillTweakMs
	}

	for _, node := range nodes {
		want := map[string]int64{}
		for _, other := range nodes {
			if other.publicKeyHex == node.publicKeyHex {
				continue
			}
			want[other.publicKeyHex] = node.hillTweakMs + other.hillTweakMs
		}
		waitForHillTweaks(t, node, want, 12*time.Second)
	}
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return wd
}

func writeConfig(t *testing.T, path string, cfg *config.NodeConfig) {
	t.Helper()
	bs, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(path, bs, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func setPeers(t *testing.T, path string, peers []string) string {
	t.Helper()
	bs, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	cfg := &config.NodeConfig{}
	if err := json.Unmarshal(bs, cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	cfg.Peers = peers
	writeConfig(t, path, cfg)
	return path
}

func reservePorts(t *testing.T, count int) []int {
	t.Helper()
	ports := make([]int, 0, count)
	for i := 0; i < count; i++ {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen: %v", err)
		}
		port := ln.Addr().(*net.TCPAddr).Port
		_ = ln.Close()
		ports = append(ports, port)
	}
	return ports
}

func startNode(t *testing.T, repoRoot, bin string, node *testNode) {
	t.Helper()
	cmd := exec.Command(bin, "-useconffile", node.cfgPath)
	cmd.Dir = repoRoot
	cmd.Stdout = node.logs
	cmd.Stderr = node.logs
	if err := cmd.Start(); err != nil {
		t.Fatalf("start %s: %v\n%s", node.name, err, node.logs.String())
	}
	node.cmd = cmd
}

func stopNode(t *testing.T, node *testNode) {
	t.Helper()
	if node == nil || node.cmd == nil || node.cmd.Process == nil {
		return
	}
	_ = node.cmd.Process.Signal(os.Interrupt)
	done := make(chan struct{})
	go func() {
		_ = node.cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = node.cmd.Process.Kill()
	}
}

func waitForAdmin(t *testing.T, sock string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(sock); err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("admin socket not ready: %s", sock)
}

func waitForHillTweaks(t *testing.T, node *testNode, want map[string]int64, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		peers, err := getPeers(node.adminSock)
		if err == nil {
			ok := true
			for key, expected := range want {
				found := false
				for _, p := range peers {
					if p.PublicKey != key {
						continue
					}
					found = true
					if p.HillTweakMs != expected {
						ok = false
					}
					break
				}
				if !found {
					ok = false
				}
			}
			if ok {
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("hillTweak mismatch for %s\nlogs:\n%s", node.name, node.logs.String())
}

func getPeers(sock string) ([]admin.PeerEntry, error) {
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := admin.AdminSocketRequest{Name: "getPeers"}
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)
	if err := enc.Encode(&req); err != nil {
		return nil, err
	}
	var resp admin.AdminSocketResponse
	if err := dec.Decode(&resp); err != nil {
		return nil, err
	}
	if resp.Status == "error" {
		return nil, fmt.Errorf("admin error: %s", resp.Error)
	}
	var peersResp admin.GetPeersResponse
	if err := json.Unmarshal(resp.Response, &peersResp); err != nil {
		return nil, err
	}
	return peersResp.Peers, nil
}
