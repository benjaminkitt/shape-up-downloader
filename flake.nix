{
  description = "Shape Up book downloader";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        # Package definition
        packages.default = pkgs.buildGoModule {
          pname = "shape-up-downloader";
          version = "1.0.0";
          src = ./.;
          vendorHash = null;
        };

        # Development environment (your existing config)
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            go-tools
            golangci-lint
            delve
            goreleaser
          ];

          shellHook = ''
            echo "Shape Up downloader development environment"
            echo "Available tools:"
            echo " * go - The Go compiler and tools"
            echo " * gopls - Go language server"
            echo " * golangci-lint - Linter"
            echo " * delve - Debugger"
            echo " * goreleaser - Release automation"
          '';
        };
      });
}
