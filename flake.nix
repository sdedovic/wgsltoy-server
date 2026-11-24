{
  description = "WGSL Toy";

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
        nodejs = pkgs.nodejs_26;
        buildGoModule = pkgs.buildGoModule.override {go = pkgs.go_1_27;};

        # the Go binary alone; it reads templates and static assets from the directories it is configured with
        server = buildGoModule {
          pname = "wgsltoy-server";
          version = "0.0.0";
          src = pkgs.lib.fileset.toSource {
            root = ./server;
            fileset = pkgs.lib.fileset.unions [./server/go.mod ./server/go.sum ./server/cmd ./server/internal];
          };
          vendorHash = "sha256-gek5aJpxSs1Yj2qChegb4ZWLD0nVa5NlOrTvF6sISV8=";
          subPackages = ["cmd/wgsltoy-server"];
        };

        # everything the browser receives: templates/ and the Vite build in dist/
        web = pkgs.buildNpmPackage {
          pname = "wgsltoy-web";
          version = "0.0.0";
          src = ./web;
          inherit nodejs;
          npmDepsHash = "sha256-+cnYsaZSj7pK4AO7Vlo8UBkiJn+SXmhO5rhvsRGuzvU=";

          npmBuildScript = "build";
          installPhase = ''
            mkdir -p $out
            cp -r dist $out/dist
            cp -r templates $out/templates
          '';
        };

        # the server configured to run against web
        bundle = pkgs.runCommand "wgsltoy" {nativeBuildInputs = [pkgs.makeWrapper];} ''
          mkdir -p $out/share/wgsltoy
          cp ${./server/application.yaml} $out/share/wgsltoy/application.yaml
          makeWrapper ${server}/bin/wgsltoy-server $out/bin/wgsltoy-server \
            --chdir $out/share/wgsltoy \
            --set-default TEMPLATES_DIR ${web}/templates \
            --set-default STATIC_DIR ${web}/dist
        '';
      in {
        # nix-fmt doesn't ignore stuff in .gitingore and thus is too slow with direnv
        formatter = pkgs.alejandra;

        packages = {
          inherit server web bundle;
          default = bundle;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go_1_27
            pkgs.just
            pkgs.go-migrate
            pkgs.golangci-lint
            nodejs
            pkgs.watchexec
          ];
        };
      }
    );
}
