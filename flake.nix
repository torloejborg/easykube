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
      lib = pkgs.lib;
      go_compiler_flags = "-tags=remote,containers_remote,exclude_graphdriver_btrfs,exclude_graphdriver_devicemapper,exclude_graphdriver_overlay,exclude_graphdriver_zfs";

        # Common packages used in all shells
            commonPackages = with pkgs; [
              (ruby.withPackages (ps: with ps; [ rouge  ]))
              jq
              yq
              gnumake
              glibcLocales
              pkg-config
              stdenv.cc
              gpgme
              libgpg-error
            ] ++ [
              pkgsUnstable.upx
              pkgsUnstable.mockgen
              pkgsUnstable.kubectl
              pkgsUnstable.kubernetes-helm
              pkgsUnstable.kustomize
              pkgsUnstable.go_1_24
            ] ;
    in {
      packages.${system}.default = pkgsUnstable.buildGoModule {
        pname = "easykube";
        version = "latest";
        src = self;
         nativeBuildInputs = [
           pkgs.pkg-config
           pkgs.stdenv.cc
         ];

         buildInputs = [
            pkgs.gpgme
            pkgs.libgpg-error
         ];

          tags = [
            "remote"
            "containers_remote"
            "exclude_graphdriver_btrfs"
            "exclude_graphdriver_devicemapper"
            "exclude_graphdriver_overlay"
            "exclude_graphdriver_zfs"
          ];

        #vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        vendorHash = "sha256-DuRFwAZzu7XXXRFrxFM2t2RH48HKKw/th0n+tVzVGBU=";
      };

      devShells.${system} = {
        default = pkgs.mkShell {

          packages = commonPackages ++  [
            (self.packages.${system}.default)
          ];

          shell = pkgs.zsh;
          impureEnv = true;
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

        light = pkgs.mkShell {
          GOFLAGS = go_compiler_flags;
          packages = commonPackages;
          shell = pkgs.zsh;
          impureEnv = true;
          shellHook = ''
            export LC_ALL=C.UTF-8
            export LANG=C.UTF-8
            export PS1="[ek-light]> "
            echo "Welcome to the easykube light dev shell (no build)"
          '';
        };
      };
    };
}
