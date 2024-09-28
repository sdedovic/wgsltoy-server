{
  description = "WGSL Toy (Server)";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = {
    self,
    flake-utils,
    nixpkgs,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };
      in {
        # nix-fmt doesn't ignore stuff in .gitingore and thus is too slow with direnv
        formatter = pkgs.alejandra;

        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
          ];
        };
      }
    );
}
