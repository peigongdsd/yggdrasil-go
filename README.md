# Yggdrasil

[![Build status](https://github.com/yggdrasil-network/yggdrasil-go/actions/workflows/ci.yml/badge.svg)](https://github.com/yggdrasil-network/yggdrasil-go/actions/workflows/ci.yml)

## Introduction

Yggdrasil is an early-stage implementation of a fully end-to-end encrypted IPv6
network. It is lightweight, self-arranging, supported on multiple platforms and
allows pretty much any IPv6-capable application to communicate securely with
other Yggdrasil nodes. Yggdrasil does not require you to have IPv6 Internet
connectivity - it also works over IPv4.

## Supported Platforms

Yggdrasil works on a number of platforms, including Linux, macOS, Ubiquiti
EdgeRouter, VyOS, Windows, FreeBSD, OpenBSD and OpenWrt.

Please see our [Installation](https://yggdrasil-network.github.io/installation.html)
page for more information. You may also find other platform-specific wrappers, scripts
or tools in the `contrib` folder.

## Building

If you want to build from source, as opposed to installing one of the pre-built
packages:

1. Install [Go](https://golang.org) (requires Go 1.22 or later)
2. Clone this repository
2. Run `./build`

Note that you can cross-compile for other platforms and architectures by
specifying the `GOOS` and `GOARCH` environment variables, e.g. `GOOS=windows
./build` or `GOOS=linux GOARCH=mipsle ./build`.

## Running

### Generate configuration

To generate static configuration, either generate a HJSON file (human-friendly,
complete with comments):

```
./yggdrasil -genconf > /path/to/yggdrasil.conf
```

... or generate a plain JSON file (which is easy to manipulate
programmatically):

```
./yggdrasil -genconf -json > /path/to/yggdrasil.conf
```

You will need to edit the `yggdrasil.conf` file to add or remove peers, modify
other configuration such as listen addresses or multicast addresses, etc.

### hillTweak (latency bias)

Yggdrasil supports a per-node `HillTweakMs` configuration value (milliseconds)
that is exchanged during the peering handshake. Each side contributes its own
value, and the sum is applied as an additive bias to the link cost used for
routing decisions. The default is `0`, which preserves existing behavior.

### Org-signed certificate authentication

Yggdrasil supports optional organization-signed authentication. An organization
can sign a node’s Yggdrasil public key using an Ed25519 “org master” key. If a
node presents a valid org-signed certificate during handshake, the peer will
accept the connection **regardless of password mismatch** (pinned keys and
AllowedPublicKeys still apply).

Config fields (hex-encoded):

```
"OrgPubKey": "<org master public key hex>"
"OrgCert": "<org-signed cert blob hex>"
```

Certificate format (v1):

```
version(1) | nodePubKey(32) | issuedAt(8) | expiresAt(8) | signature(64)
```

Signature is Ed25519 over the payload (version..expiresAt). `expiresAt = 0`
means no expiry. If `OrgPubKey` is not set or `OrgCert` is missing/invalid,
normal password-based authentication applies.

Key generation (organization):

```
yggdrasilctl orgKeygen > /etc/yggdrasil/org.key
```

This writes a file containing both keys:

```
OrgPubKey: <hex>
OrgPrivKey: <hex>
```

Distribute `OrgPubKey` to all nodes (in config). Keep `OrgPrivKey` private.

Signing a node:

1) Get the node's public key (from its config or `yggdrasil -publickey`).
2) Generate an org-signed cert:

```
yggdrasilctl orgSign pubkey=<node_pubkey_hex> orgkey=/path/to/org.key
```

`org.key` should contain a hex-encoded Ed25519 private key (64 bytes) or seed (32 bytes).
By default, certs never expire. Optional fields:

```
yggdrasilctl orgSign pubkey=<hex> orgkey=/path/to/org.key issued_at=<unix> expires_at=<unix>
```

You can generate an org keypair with:

```
yggdrasilctl orgKeygen
```

To output as JSON:

```
yggdrasilctl -json orgKeygen
```

Config example (node):

```
{
  "OrgPubKey": "<org master public key hex>",
  "OrgCert": "<org-signed cert blob hex>",
  "Peers": [
    "tls://peer.example.com:8091"
  ]
}
```

### Run Yggdrasil

To run with the generated static configuration:

```
./yggdrasil -useconffile /path/to/yggdrasil.conf
```

To run in auto-configuration mode (which will use sane defaults and random keys
at each startup, instead of using a static configuration file):

```
./yggdrasil -autoconf
```

You will likely need to run Yggdrasil as a privileged user or under `sudo`,
unless you have permission to create TUN/TAP adapters. On Linux this can be done
by giving the Yggdrasil binary the `CAP_NET_ADMIN` capability.

## Documentation

Documentation is available [on our website](https://yggdrasil-network.github.io).

- [Installing Yggdrasil](https://yggdrasil-network.github.io/installation.html)
- [Configuring Yggdrasil](https://yggdrasil-network.github.io/configuration.html)
- [Frequently asked questions](https://yggdrasil-network.github.io/faq.html)
- [Version changelog](CHANGELOG.md)

## NixOS / Flake Usage

This fork ships a Nix flake with both a package and a NixOS module.

Example `flake.nix` snippet:

```
{
  inputs.yggdrasil.url = "github:peigongdsd/yggdrasil-go";

  outputs = { self, nixpkgs, yggdrasil, ... }:
  let
    system = "x86_64-linux";
  in {
    nixosConfigurations.host = nixpkgs.lib.nixosSystem {
      inherit system;
      modules = [
        yggdrasil.nixosModules.yggdrasil
        ({ config, pkgs, ... }: {
          nixpkgs.overlays = [ yggdrasil.overlays.default ];
          services.yggdrasil = {
            enable = true;
            settings = {
              HillTweakMs = 250;
              Listen = [ "tcp://0.0.0.0:12345" ];
              Peers = [ "tcp://1.2.3.4:12345" ];
            };
          };
        })
      ];
    };
  };
}
```

This overlay provides `pkgs.yggdrasil` from the fork and the module is available
at `yggdrasil.nixosModules.yggdrasil`.

## Communities

A number of IRC communities exist, including the `#yggdrasil` IRC channel on [libera.chat](https://libera.chat) and various others on [Yggdrasil-internal IRC networks](https://yggdrasil-network.github.io/services.html#irc).

## License

This code is released under the terms of the LGPLv3, but with an added exception
that was shamelessly taken from [godeb](https://github.com/niemeyer/godeb).
Under certain circumstances, this exception permits distribution of binaries
that are (statically or dynamically) linked with this code, without requiring
the distribution of Minimal Corresponding Source or Minimal Application Code.
For more details, see: [LICENSE](LICENSE).
