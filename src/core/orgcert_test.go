package core

import (
	"crypto/ed25519"
	"testing"
	"time"
)

func TestOrgCertVerify(t *testing.T) {
	orgPub, orgPriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("org keygen: %v", err)
	}
	nodePub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("node keygen: %v", err)
	}
	now := time.Now().Unix()
	cert, err := EncodeOrgCertV1(nodePub, now, now+3600, orgPriv)
	if err != nil {
		t.Fatalf("encode cert: %v", err)
	}
	if err := VerifyOrgCertV1(cert, orgPub, nodePub); err != nil {
		t.Fatalf("verify cert: %v", err)
	}
}
