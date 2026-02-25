{
  description = "Yggdrasil-go fork with hillTweak support";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        lib = pkgs.lib;
        version = "0.5.13";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "yggdrasil";
          inherit version;

          src = pkgs.fetchFromGitHub {
            owner = "peigongdsd";
            repo = "yggdrasil-go";
            rev = "v${version}";
            hash = "sha256-NlNQnYmK//p35pj2MInD6RVsajM/bGDhOuzOZZYoWRw=";
          };

          vendorHash = lib.fakeSha256;

          subPackages = [
            "cmd/genkeys"
            "cmd/yggdrasil"
            "cmd/yggdrasilctl"
          ];

          ldflags = [
            "-X github.com/yggdrasil-network/yggdrasil-go/src/version.buildVersion=${version}"
            "-X github.com/yggdrasil-network/yggdrasil-go/src/version.buildName=yggdrasil"
            "-X github.com/yggdrasil-network/yggdrasil-go/src/config.defaultAdminListen=unix:///var/run/yggdrasil/yggdrasil.sock"
            "-s"
            "-w"
          ];

          passthru.tests.basic = pkgs.nixosTests.yggdrasil;

          meta = with lib; {
            description = "Experiment in scalable routing as an encrypted IPv6 overlay network";
            homepage = "https://yggdrasil-network.github.io/";
            license = licenses.lgpl3;
            mainProgram = "yggdrasil";
            maintainers = with maintainers; [
              gazally
              lassulus
              peigongdsd
            ];
          };
        };
      }) // {
        nixosModules.yggdrasil = import ./nixos-module.nix;
      };
}
