{
  description = "TaserudConsulting/goprocmgr";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    flake-utils,
    nixpkgs,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
    in {
      packages = flake-utils.lib.flattenTree {
        default = pkgs.buildGoModule (let
          versionTag = "0.0.0";
          version = "${versionTag}.${nixpkgs.lib.substring 0 8 self.lastModifiedDate}.${self.shortRev or "dirty"}";
        in {
          pname = "goprocmgr";
          inherit version;

          prePatch = ''
            substituteInPlace main.go --replace "%undefined-version%" ${versionTag}
          '';

          src = ./.;

          vendorHash = "sha256-aA+FeMBLhvh4pg3W4eHEfBtOf5oUnDDUkKnHEQA/+vI=";
        });
      };

      devShells = flake-utils.lib.flattenTree {
        default = pkgs.mkShell {
          buildInputs = [
            pkgs.gnumake
            pkgs.delve # debugging
            pkgs.go # language
            pkgs.gopls # language server
          ];
        };
      };

      formatter = pkgs.alejandra;
    });
}
