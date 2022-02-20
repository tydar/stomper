with import <nixpkgs> {};

buildGoModule {
        pname = "stomper";
        version = "0.0.1";

        src = ./.;

        # requires running go mod vendor prior to build
        # on NixOS, do this:
        # $ nix-shell -p go
        # $ go mod vendor
        # $ nix build

        vendorSha256 = null;
}
