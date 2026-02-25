package core

import (
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"time"
)

const orgCertV1Size = 1 + ed25519.PublicKeySize + 8 + 8 + ed25519.SignatureSize

var (
	ErrOrgCertInvalidLength  = errors.New("org cert invalid length")
	ErrOrgCertInvalidVersion = errors.New("org cert invalid version")
	ErrOrgCertKeyMismatch    = errors.New("org cert public key mismatch")
	ErrOrgCertExpired        = errors.New("org cert expired")
	ErrOrgCertVerifyFailed   = errors.New("org cert verification failed")
)

func EncodeOrgCertV1(nodePub ed25519.PublicKey, issuedAt, expiresAt int64, orgPriv ed25519.PrivateKey) ([]byte, error) {
	if len(nodePub) != ed25519.PublicKeySize {
		return nil, ErrOrgCertKeyMismatch
	}
	payload := make([]byte, 1+ed25519.PublicKeySize+8+8)
	payload[0] = 1
	copy(payload[1:1+ed25519.PublicKeySize], nodePub)
	binary.BigEndian.PutUint64(payload[1+ed25519.PublicKeySize:], uint64(issuedAt))
	binary.BigEndian.PutUint64(payload[1+ed25519.PublicKeySize+8:], uint64(expiresAt))
	sig := ed25519.Sign(orgPriv, payload)
	return append(payload, sig...), nil
}

func VerifyOrgCertV1(cert []byte, orgPub ed25519.PublicKey, expectedNodePub ed25519.PublicKey) error {
	if len(cert) != orgCertV1Size {
		return ErrOrgCertInvalidLength
	}
	if cert[0] != 1 {
		return ErrOrgCertInvalidVersion
	}
	payload := cert[:1+ed25519.PublicKeySize+8+8]
	sig := cert[len(cert)-ed25519.SignatureSize:]
	nodePub := payload[1 : 1+ed25519.PublicKeySize]
	if len(expectedNodePub) == ed25519.PublicKeySize && !ed25519.PublicKey(nodePub).Equal(expectedNodePub) {
		return ErrOrgCertKeyMismatch
	}
	issuedAt := int64(binary.BigEndian.Uint64(payload[1+ed25519.PublicKeySize:]))
	expiresAt := int64(binary.BigEndian.Uint64(payload[1+ed25519.PublicKeySize+8:]))
	_ = issuedAt
	if expiresAt > 0 && time.Now().Unix() > expiresAt {
		return ErrOrgCertExpired
	}
	if !ed25519.Verify(orgPub, payload, sig) {
		return ErrOrgCertVerifyFailed
	}
	return nil
}
