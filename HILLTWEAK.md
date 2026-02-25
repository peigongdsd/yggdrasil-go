# HillTweak: Biasing Link Cost to Prefer or Avoid Routes

This document explains **HillTweak**, why it exists, and how to use it.
It is written for operators who want **fine‑grained control over routing
preferences** without changing the routing algorithm itself.

---

## What HillTweak is

HillTweak is a **per‑node, per‑link cost bias** that is added to the measured
latency when Yggdrasil computes routing cost. It lets you make some nodes
**less likely to be selected** as next hops without disconnecting them or
changing the topology.

In short:

> **Effective link cost = measured RTT + HillTweak**

This is applied per connection, so the bias affects every peer link involving
that node.

---

## Why use it

Common use cases:

- **Keep “backup” nodes reachable but unlikely to be used**
- **De‑prefer relay nodes** while keeping them available
- **Steer traffic away from sensitive or overloaded nodes**

---

## How it works (precise behavior)

Each node publishes its own `HillTweakMs` in its config. During handshake,
peers exchange their values. The effective tweak for a link is the **sum** of
both sides:

```
effective_hill_tweak = local_hill_tweak + remote_hill_tweak
```

That sum is added to the measured RTT when computing routing cost.

Example:

- Node A sets `HillTweakMs = 250`
- Node B sets `HillTweakMs = 400`

The link between A and B behaves as if its RTT were **650ms higher**, making it
less attractive for routing decisions.

---

## How to configure it

Add this to your node’s config:

```json
{
  "HillTweakMs": 250
}
```

If you’re using the NixOS module, set:

```nix
services.yggdrasil.settings.HillTweakMs = 250;
```

---

## How to verify it’s working

Use JSON output:

```
yggdrasilctl -json getPeers
```

Look for the field:

```
"hill_tweak_ms": 250
```

The value shown is the **effective sum** (local + remote), not just your local
setting.

---

## Notes and cautions

- HillTweak does **not** block routes; it only biases them.
- Very large values can make a node almost never selected for routing.
- Use it sparingly to avoid unintentionally fragmenting the mesh.

---

## Summary

HillTweak is a **soft routing policy** tool:

- **Set it high** on nodes you want to be **rarely used**.
- **Leave it at 0** on normal nodes.
- The network stays connected, but routing preference shifts away from the
  tweaked nodes.
