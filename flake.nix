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
        pkgsUnstable.go_1_25
        pkgsUnstable.gpgme
      ];
    in {
      packages.${system}.default = pkgsUnstable.buildGoModule {
        pname = "easykube";
        version = "latest";
        src = self;

        vendorHash = "sha256-9+aRA8yzZ84DApX4S5WEddL90XFQvrzLKLL5B2cjG4c=";

        tags = go_tags;

        # Add version information similar to Makefile
        ldflags = [ "-X github.com/torloejborg/easykube/pkg/vars.Version=${self.lastModifiedDate or "unknown"}" ];

        # Add build dependencies for storage drivers
        buildInputs = with pkgsUnstable; [
          btrfs-progs
          lvm2
        ];

        # Set CGO flags for btrfs
        CGO_CFLAGS = "-I${pkgsUnstable.btrfs-progs}/include";
        CGO_LDFLAGS = "-L${pkgsUnstable.btrfs-progs}/lib";
      };

      apps.${system}.default = {
        type = "app";
        program = "${self.packages.${system}.default}/bin/easykube";
      };

      devShells.${system} = {
        default = pkgs.mkShell {
          inputsFrom = [
            self.packages.${system}.default
          ];

          packages = commonPackages;

          shell = pkgs.zsh;
          impureEnv = true;

          shellHook = ''
            export LC_ALL=C.UTF-8
            export LANG=C.UTF-8
            export PS1="[ek-dev]> "
            echo "easykube has been built and is available on PATH"
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
