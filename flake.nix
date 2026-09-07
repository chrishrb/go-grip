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
      default = pkgs.buildGoModule rec {
        pname = "go-grip";
        version = self.shortRev or self.dirtyShortRev or "dev";
        src = self;
        ldflags = [
          "-s"
          "-w"
          "-X"
          "github.com/chrishrb/go-grip/cmd.version=${version}"
        ];
        # Only for updating vendorHas
        # vendorHash = "sha256-RRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRR=";
        vendorHash = "sha256-QsLiCsFY6nI85jsEZtAgmObEKpBSZWhzZk+TlukM8JU=";
      };
    });
    devShells = forAllSystems (pkgs: {
      default = self.packages.${pkgs.system}.default;
    });
  };
}
