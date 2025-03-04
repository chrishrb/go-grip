{
  description = "go-grip - render your markdown files local";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };
  outputs = {
    self,
    nixpkgs,
    ...
  }: let
    systems = [
      "aarch64-linux"
      "aarch64-darwin"
      "x86_64-darwin"
      "x86_64-linux"
    ];
    forAllSystems = f:
      nixpkgs.lib.genAttrs systems (system: let
        pkgs = import nixpkgs {inherit system;};
      in
        f pkgs);
  in {
    packages = forAllSystems (pkgs: {
      default = pkgs.buildGoModule {
        name = "go-grip";
        src = self;
        # Only for updating vendorHas
        # vendorHash = "sha256-RRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRR=";
        vendorHash = "sha256-aU6vo/uqJzctD7Q8HPFzHXVVJwMmlzQXhAA6LSkRAow=";
      };
    });
    devShells = forAllSystems (pkgs: {
      default = self.packages.${pkgs.system}.default;
    });
  };
}
