{
    inputs = {
        nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
        flake-utils.url = "github:numtide/flake-utils";
        gomod2nix.url = "github:tweag/gomod2nix";
        gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
        gomod2nix.inputs.utils.follows = "flake-utils";
    };

    outputs = { self, nixpkgs, flake-utils, gomod2nix }:
        let
            out = system:
                let
                    pkgs = import nixpkgs {
                        inherit system;
                        overlays = [ gomod2nix.overlay ];
                    };
                in {
                        devShell = pkgs.mkShell { buildInputs = with pkgs; [ go ]; };

                        defaultPackage = pkgs.buildGoApplication {
                            pname = "stomper";
                            version = "latest";
                            goPackagePath = "github.com/tydar/stomper";
                            src = ./.;
                            modules = ./gomod2nix.toml;
                        };

                        defaultApp =
                            flake-utils.lib.mkApp { drv = self.defaultPackage."${system}"; };
                };
            in with flake-utils.lib; eachSystem defaultSystems out;
}
