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
          vendorHash = "sha256-tfvhMbe0uSWIfaUUawEYe+7ckBttwM1IokKAWBLi8ig=";
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
