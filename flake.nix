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
        version = "latest";
        src = self;

        vendorHash = "sha256-K3R8blmcMf67ztFS4TbpnrqVHhjotX0jRiWXttfdJSE=";
      };

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          jq
          yq
          pkgs.gnumake
          pkgs.glibcLocales
          (self.packages.${system}.default)
        ] ++ [
          pkgsUnstable.upx
          pkgsUnstable.mockgen
          pkgsUnstable.kubectl
          pkgsUnstable.kubernetes-helm
          pkgsUnstable.kustomize
          pkgsUnstable.go_1_24
        ];

        shell = pkgs.zsh;
        impureEnv = true;  # Allow access to system environment variables
        shellHook = ''
              export LC_ALL=C.UTF-8
              export LANG=C.UTF-8
              export PS1="[ek-dev]> "
              source <(easykube completion bash)

              echo "Welcome to the easykube dev shell"
              echo
              easykube
        '';
      };
    };
}
