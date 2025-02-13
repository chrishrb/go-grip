{
  description = "go-grip - render your markdown files local";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, ... }@inputs: inputs.flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        go-grip = pkgs.buildGoModule {
          name = "go-grip";
          src = self;
          # Only for updating vendorHas
          # vendorHash = "sha256-RRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRR=";
          vendorHash = "sha256-aU6vo/uqJzctD7Q8HPFzHXVVJwMmlzQXhAA6LSkRAow=";
        };
      in
      {
        packages.default = go-grip;
        devShells.default = pkgs.mkShell {
          nativeBuildInputs = [ go-grip ];
        };
      }
    );
}
