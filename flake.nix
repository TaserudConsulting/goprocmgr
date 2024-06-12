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
          versionTag = "1.1.0";
          pname = "goprocmgr";
          version = "${versionTag}.${nixpkgs.lib.substring 0 8 self.lastModifiedDate}.${self.shortRev or "dirty"}";
        in {
          inherit pname version;

          nativeBuildInputs = [
            pkgs.pandoc
          ];

          prePatch = ''
            substituteInPlace CLI.md main.go --replace-fail "%undefined-version%" ${versionTag}
          '';

          postBuild = ''
            pandoc -s -t man CLI.md -o goprocmgr.1
          '';

          postInstall = ''
            install -Dm644 goprocmgr.1 $out/share/man/man1/goprocmgr.1
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
