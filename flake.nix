{
  description = "Yggdrasil-go fork with hillTweak support";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      overlay = final: prev: {
        yggdrasil = self.packages.${final.system}.default;
      };
    in
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
            rev = "f939c7f791f3c76cd216db997e5003ae1c0d8d74";
            hash = "sha256-paFee/MLH9OUTzJoX+bAsZgizdb+CmBNA4IBQCHnqvU=";
          };

          vendorHash = "sha256-oEViEh3oUbhzpn8uyasImFUarfMJpgO+VdCvoCctEnY=";

          subPackages = [
            "cmd/genkeys"
            "cmd/yggdrasil"
            "cmd/yggdrasilctl"
          ];

          ldflags = [
            "-X github.com/yggdrasil-network/yggdrasil-go/src/version.buildVersion=${version}+hilltweak.${"20260225173418"}"
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
        overlays.default = overlay;
        nixosModules.yggdrasil = import ./nixos-module.nix;
      };
}
