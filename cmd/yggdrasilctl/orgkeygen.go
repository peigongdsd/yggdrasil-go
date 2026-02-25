package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func runOrgKeygen(asJSON bool) error {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	if asJSON {
		out := map[string]string{
			"org_pub_key":  hex.EncodeToString(pub),
			"org_priv_key": hex.EncodeToString(priv),
		}
		enc := json.NewEncoder(stdoutWriter{})
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}
	fmt.Println("OrgPubKey:", hex.EncodeToString(pub))
	fmt.Println("OrgPrivKey:", hex.EncodeToString(priv))
	return nil
}

type stdoutWriter struct{}

func (stdoutWriter) Write(p []byte) (int, error) {
	return fmt.Print(string(p))
}
