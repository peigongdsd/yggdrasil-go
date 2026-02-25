package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"crypto/ed25519"

	"github.com/yggdrasil-network/yggdrasil-go/src/core"
)

func runOrgSign(args []string) error {
	kv := map[string]string{}
	for _, a := range args {
		tokens := strings.SplitN(a, "=", 2)
		if len(tokens) != 2 {
			continue
		}
		kv[tokens[0]] = tokens[1]
	}
	pubHex := kv["pubkey"]
	orgKeyPath := kv["orgkey"]
	if pubHex == "" || orgKeyPath == "" {
		return errors.New("usage: orgSign pubkey=<hex> orgkey=/path/to/org.key [issued_at=unix] [expires_at=unix]")
	}
	pub, err := hex.DecodeString(pubHex)
	if err != nil {
		return fmt.Errorf("decode pubkey: %w", err)
	}
	orgKeyBytes, err := os.ReadFile(orgKeyPath)
	if err != nil {
		return fmt.Errorf("read orgkey: %w", err)
	}
	orgKeyHex := strings.TrimSpace(string(orgKeyBytes))
	orgKeyRaw, err := hex.DecodeString(orgKeyHex)
	if err != nil {
		return fmt.Errorf("decode orgkey hex: %w", err)
	}
	var orgPriv ed25519.PrivateKey
	switch len(orgKeyRaw) {
	case ed25519.PrivateKeySize:
		orgPriv = ed25519.PrivateKey(orgKeyRaw)
	case ed25519.SeedSize:
		orgPriv = ed25519.NewKeyFromSeed(orgKeyRaw)
	default:
		return fmt.Errorf("orgkey length must be %d or %d bytes", ed25519.PrivateKeySize, ed25519.SeedSize)
	}

	issuedAt := time.Now().Unix()
	expiresAt := int64(0)
	if v := kv["issued_at"]; v != "" {
		if issuedAt, err = strconv.ParseInt(v, 10, 64); err != nil {
			return fmt.Errorf("issued_at: %w", err)
		}
	}
	if v := kv["expires_at"]; v != "" {
		if expiresAt, err = strconv.ParseInt(v, 10, 64); err != nil {
			return fmt.Errorf("expires_at: %w", err)
		}
	}

	cert, err := core.EncodeOrgCertV1(ed25519.PublicKey(pub), issuedAt, expiresAt, orgPriv)
	if err != nil {
		return fmt.Errorf("encode org cert: %w", err)
	}
	fmt.Println(hex.EncodeToString(cert))
	return nil
}
