{
  description = "easykube";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, unstable }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      pkgsUnstable = unstable.legacyPackages.${system};
    in {
      packages.${system}.default = pkgsUnstable.buildGoModule {
        pname = "easykube";
        version = "1.1.5";
        src = self;

        vendorHash = "sha256-XA0kCP+pe1ZmsOdjT/HRUi5XzDg0/yEz0EupKVL/GQg="; # will be filled after first build
      };

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          kubectl
          helm
          kustomize
          jq
          yq
          (self.packages.${system}.default)
        ] ++ [
          pkgsUnstable.go_1_24
        ];
      };
    };
}
