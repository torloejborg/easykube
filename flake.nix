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

        vendorHash = "sha256-XA0kCP+pe1ZmsOdjT/HRUi5XzDg0/yEz0EupKVL/GQg=";
      };

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          jq
          yq
          zsh
          (self.packages.${system}.default)
        ] ++ [
          pkgsUnstable.kubectl
          pkgsUnstable.kubernetes-helm
          pkgsUnstable.kustomize
          pkgsUnstable.go_1_24
        ];

        shell = pkgs.zsh;

        shellHook = ''

          # Aliases
          alias k="kubectl"
          alias h="helm"
          alias ek="easykube"

          echo "Welcome to easykube dev shell!"
          echo "Run 'easykube --help' to get started"
        '';
      };


    };
}
