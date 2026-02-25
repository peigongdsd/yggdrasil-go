package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/yggdrasil-network/yggdrasil-go/src/core"
)

func TestOrgSignOutputsValidCert(t *testing.T) {
	orgPub, orgPriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("org keygen: %v", err)
	}
	nodePub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("node keygen: %v", err)
	}

	tmp := t.TempDir()
	orgKeyPath := tmp + "/org.key"
	if err := os.WriteFile(orgKeyPath, []byte(hex.EncodeToString(orgPriv)), 0o600); err != nil {
		t.Fatalf("write org key: %v", err)
	}

	out, err := captureStdout(func() error {
		return runOrgSign([]string{
			"pubkey=" + hex.EncodeToString(nodePub),
			"orgkey=" + orgKeyPath,
		})
	})
	if err != nil {
		t.Fatalf("runOrgSign: %v", err)
	}

	hexCert := strings.TrimSpace(string(out))
	cert, err := hex.DecodeString(hexCert)
	if err != nil {
		t.Fatalf("decode cert hex: %v", err)
	}
	if err := core.VerifyOrgCertV1(cert, orgPub, nodePub); err != nil {
		t.Fatalf("verify cert: %v", err)
	}
}
