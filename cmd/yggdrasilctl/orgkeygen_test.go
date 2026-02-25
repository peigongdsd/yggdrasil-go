package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestOrgKeygenJSON(t *testing.T) {
	out, err := captureStdout(func() error {
		return runOrgKeygen(true)
	})
	if err != nil {
		t.Fatalf("runOrgKeygen: %v", err)
	}
	var payload map[string]string
	if err := json.Unmarshal(bytes.TrimSpace(out), &payload); err != nil {
		t.Fatalf("json parse: %v\n%s", err, string(out))
	}
	pub := payload["org_pub_key"]
	priv := payload["org_priv_key"]
	if pub == "" || priv == "" {
		t.Fatalf("missing keys in output: %v", payload)
	}
	if _, err := hex.DecodeString(pub); err != nil {
		t.Fatalf("pub hex decode: %v", err)
	}
	if _, err := hex.DecodeString(priv); err != nil {
		t.Fatalf("priv hex decode: %v", err)
	}
}

func captureStdout(fn func() error) ([]byte, error) {
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	os.Stdout = w
	err = fn()
	_ = w.Close()
	os.Stdout = orig
	if err != nil {
		_, _ = r.Read(make([]byte, 1024))
		_ = r.Close()
		return nil, err
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	_ = r.Close()
	return buf.Bytes(), nil
}
