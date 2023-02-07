{
  description = "Report unwanted import path and declaration usages";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        version = "1.11.0";
        pkgs = import nixpkgs {
          inherit system;
          config = {allowUnfree = true;};
        };
      in rec {
        packages.faillint = pkgs.buildGoModule {
          pname = "faillint";
          inherit version;
          src = self;
          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
            "-X main.date=${self.lastModifiedDate}"
          ];
          CGO_ENABLED = false;

          vendorSha256 = "5OR6Ylkx8AnDdtweY1B9OEcIIGWsY8IwTHbR/LGnqFI=";
        };

        devShell = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go_1_19
            golangci-lint
            gotools
          ];
        };

        formatter = pkgs.alejandra;

        defaultPackage = packages.faillint;
        apps.faillint = flake-utils.lib.mkApp {drv = packages.faillint;};
        defaultApp = apps.faillint;
      }
    );
}
