# Organization-Based Authentication (Org-Signed Certs)

This document describes how to use the organization-signed authentication
feature in this fork of Yggdrasil.

## Overview

An organization generates a master Ed25519 keypair. The master **private** key
is used to sign each node’s Yggdrasil public key, producing an **OrgCert**.
Every node is configured with the organization **public** key, and (optionally)
its own OrgCert. During handshake, if a peer presents a valid OrgCert signed by
your organization, the connection is accepted even if the password does not
match (pinned keys and AllowedPublicKeys still apply).

## 1) Generate the organization keypair

On a secure machine:

```
yggdrasilctl orgKeygen > /etc/yggdrasil/org.key
```

This creates a key file like:

```
OrgPubKey: <hex>
OrgPrivKey: <hex>
```

Keep `OrgPrivKey` secret. Distribute `OrgPubKey` to all nodes.

## 2) Get a node’s Yggdrasil public key

Option A: from an existing config file

```
yggdrasil -useconffile /path/to/yggdrasil.conf -publickey
```

Option B: from a running node (same key, if config is persistent)

```
yggdrasil -useconffile /path/to/yggdrasil.conf -publickey
```

(You can also parse the `PrivateKey` in the config, but using `-publickey`
is simplest.)

## 3) Sign the node public key

Use the org master key to produce a cert:

```
yggdrasilctl orgSign pubkey=<node_pubkey_hex> orgkey=/etc/yggdrasil/org.key
```

Optional: set explicit validity:

```
yggdrasilctl orgSign pubkey=<hex> orgkey=/etc/yggdrasil/org.key issued_at=<unix> expires_at=<unix>
```

The output is a hex blob — this is the **OrgCert**.

## 4) Configure nodes

Add these fields to each node’s config:

```
{
  "OrgPubKey": "<org master public key hex>",
  "OrgCert": "<org-signed cert blob hex>",
  ...
}
```

Notes:
- `OrgPubKey` must be present on every node that should accept org-signed peers.
- `OrgCert` is only required for nodes that should authenticate with the org cert.

## 5) Restart Yggdrasil

```
systemctl restart yggdrasil
```

## 6) Verify

```
yggdrasilctl -json getPeers
```

If org auth is in use, the connection should succeed even when passwords
mismatch (provided the cert is valid).

## Certificate Format (for reference)

```
version(1) | nodePubKey(32) | issuedAt(8) | expiresAt(8) | signature(64)
```

Signature is Ed25519 over the payload (version..expiresAt). `expiresAt = 0`
means no expiry.
