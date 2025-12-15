{
  description = "easykube";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    unstable.url = "github:NixOS/nixpkgs/nixos-unstable";

    eksrc = {
        url = "github:torloejborg/easykube/feat/podman";
        flake = false; # important for Go modules
      };

  };

  outputs = { self, nixpkgs, unstable, eksrc }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      pkgsUnstable = unstable.legacyPackages.${system};
      lib = pkgs.lib;

        # Common packages used in all shells
            commonPackages = with pkgs; [
              (ruby.withPackages (ps: with ps; [ rouge  ]))
              jq
              yq
              gnumake
              glibcLocales
            ] ++ [
              pkgsUnstable.upx
              pkgsUnstable.mockgen
              pkgsUnstable.kubectl
              pkgsUnstable.kubernetes-helm
              pkgsUnstable.kustomize
              pkgsUnstable.go_1_25
            ] ;
    in {
      packages.${system}.default = pkgsUnstable.pkgsStatic.buildGoModule {
        pname = "easykube";
        version = "latest";

        src = eksrc;

        vendorHash = "sha256-vuwzjHu0VaewO7Di70HfcHwAxdTdVq0N0+Vy3ktgX5E=";

        env.CGO_ENABLED = "0";

        ldflags = [
          "-s"
          "-w"
          "-extldflags=-static"
        ];

        meta = with lib; {
          description = "easykube - Kubernetes cluster management tool";
          license = licenses.mit;
          platforms = platforms.linux;
        };
      };

       devShells.${system} = {
       default = pkgs.mkShell {

         packages = commonPackages ++ [ self.packages.${system}.default ];

         shell = pkgs.zsh;
         impureEnv = true;
          shellHook = ''
            export LC_ALL=C.UTF-8
            export LANG=C.UTF-8
            export PS1="ek-dev$ "
            
            # Only source completion if binary exists
            if command -v easykube >/dev/null 2>&1; then
              source <(easykube completion bash)
            fi

            echo "Welcome to the easykube dev shell"
            echo
            easykube
          '';
       };
        light = pkgs.mkShell {
          packages = commonPackages;
          shell = pkgs.zsh;
          impureEnv = true;
          shellHook = ''
            export LC_ALL=C.UTF-8
            export LANG=C.UTF-8
            export PS1="ek-light $ "
            echo "Welcome to the easykube light dev shell (no build)"
          '';
        };
      };
    };
}
