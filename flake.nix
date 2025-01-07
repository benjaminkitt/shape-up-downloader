{
  description = "Shape Up book downloader development environment";

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