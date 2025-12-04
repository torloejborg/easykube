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

      go_tags = [
        "remote"
        "containers_remote"
        "exclude_graphdriver_btrfs"
        "exclude_graphdriver_devicemapper"
        "exclude_graphdriver_overlay"
        "exclude_graphdriver_zfs"
        "containers_image_openpgp"
        "containers_image_storage_stub"
      ];

      go_flags = "-tags=${builtins.concatStringsSep "," go_tags}";

      commonPackages = with pkgs; [
        (ruby.withPackages (ps: with ps; [ rouge ]))
        jq
        yq
        gnumake
      ] ++ [
        pkgsUnstable.upx
        pkgsUnstable.mockgen
        pkgsUnstable.kubectl
        pkgsUnstable.kubernetes-helm
        pkgsUnstable.kustomize
        pkgsUnstable.go_1_24
      ];
    in {
      packages.${system}.default = pkgsUnstable.buildGoModule {
        pname = "easykube";
        version = "latest";
        src = self;

        vendorHash = null;
        modSha256 = "";

        tags = go_tags;

        # ensure go runs in module mode (optional/redundant with vendorHash=null)
        buildFlags = [ "-mod=mod" ];

      };

      devShells.${system} = {
        default = pkgs.mkShell {
          packages = commonPackages;

          shell = pkgs.zsh;
          impureEnv = true;

          shellHook = ''
            export LC_ALL=C.UTF-8
            export LANG=C.UTF-8
            export PS1="[ek-dev]> "
            echo "Welcome to the easykube dev shell"
            echo
          '';
        };

        light = pkgs.mkShell {
          GOFLAGS = go_flags;

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
