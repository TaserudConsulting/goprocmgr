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
          versionTag = "1.3.0";
          pname = "goprocmgr";
          version = "${versionTag}.${nixpkgs.lib.substring 0 8 self.lastModifiedDate}.${self.shortRev or "dirty"}";
        in {
          inherit pname version;

          nativeBuildInputs = [
            pkgs.pandoc
            pkgs.installShellFiles
          ];

          prePatch = ''
            substituteInPlace CLI.md main.go --replace-fail "%undefined-version%" ${version}
          '';

          postBuild = ''
            pandoc -s -t man CLI.md -o goprocmgr.1
          '';

          postInstall = ''
            install -Dm644 goprocmgr.1 $out/share/man/man1/goprocmgr.1
            installShellCompletion --cmd goprocmgr contrib/completions/goprocmgr.{fish,bash}
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

      checks = flake-utils.lib.flattenTree {
        default = pkgs.nixosTest {
          name = "goprocmgr-integration-test";
          nodes.machine = {
            config,
            pkgs,
            ...
          }: {
            # Install package
            environment.systemPackages = [
              self.packages.${system}.default
            ];

            # Create service
            systemd.services.${self.packages.${system}.default.pname} = {
              description = self.packages.${system}.default.pname;
              after = ["network.target"];
              wantedBy = ["multi-user.target"];
              serviceConfig.ExecStart = "${self.packages.${system}.default}/bin/${self.packages.${system}.default.pname}";
              serviceConfig.Restart = "always";
            };
          };

          testScript = ''
            machine.start()
            machine.wait_for_unit("${self.packages.${system}.default.pname}.service")

            # Check version output from command line util
            version = machine.succeed("${self.packages.${system}.default.pname} -version")
            assert '${self.packages.${system}.default.pname} version ${self.packages.${system}.default.version}' in version, \
              "Version output mismatch, got: '" + version + "', expected '${self.packages.${system}.default.version}'"

            # Test connecting to the running instance
            machine.succeed("${self.packages.${system}.default.pname} -list")

            # Test adding a new server
            machine.succeed("${self.packages.${system}.default.pname} -add 'echo hello; sleep 1; echo world'")

            # Fetch the list of servers and check that the new server is in the list
            list = machine.succeed("${self.packages.${system}.default.pname} -list")
            assert 'echo hello; sleep 1; echo world' in list, "Expected 'echo hello; sleep 1; echo world' in list, got: '" + list + "'"
            assert 'tmp' in list, "Expected 'tmp' in list, got: '" + list + "'"
          '';
        };
      };

      formatter = pkgs.alejandra;
    });
}
