{
  description = "easykube";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    gomod2nix.url = "github:nix-community/gomod2nix";
  };

  outputs = { self, nixpkgs, unstable, gomod2nix}:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      pkgsUnstable = import unstable {
        inherit system;
        overlays = [ gomod2nix.overlays.default ];
      };
      lib = pkgs.lib;

      # Common packages used in all shells
      commonPackages = with pkgs; [
        (ruby.withPackages (ps: with ps; [ rouge ]))
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
      ];

       docsPackages = with pkgs; [
                    pkgsUnstable.nodejs
                    pkgsUnstable.asciidoctor
                    pkgsUnstable.pandoc
                    pkgsUnstable.antora
                    pkgsUnstable.termtosvg
                    pkgsUnstable.plantuml
                    pkgsUnstable.graphviz
                    pkgsUnstable.ruby
                 ];

    in {
      packages.${system}.default = pkgsUnstable.buildGoApplication {
        pname = "easykube";
        version = "latest";
        src = self;
        modules = ./gomod2nix.toml;
        CGO_ENABLED = 0;
        ldflags = [
          "-s"
          "-w"
          "-extldflags=-static"
        ];
      };

       devShells.${system} = {
       default = pkgs.mkShell {
       packages = commonPackages ++ [ self.packages.${system}.default ];
       shell = pkgs.zsh;
         impureEnv = true;
          shellHook = ''
            export LC_ALL=C.UTF-8
            export LANG=C.UTF-8
            export PS1="ek-dev $ "
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

        docs = pkgs.mkShell {
          packages = commonPackages ++ docsPackages;
          impureEnv = true;
          shellHook = ''
            export LC_ALL=C.UTF-8
            export LANG=C.UTF-8
            export PS1="ek-docs $ "
            echo "Welcome to the easykube doc builder shell (no build)"
          '';
        };

      };
    };
}
