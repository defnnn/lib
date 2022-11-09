{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    home.url = "path:/home/ubuntu/dev";
  };

  outputs =
    { self
    , nixpkgs
    , flake-utils
    , home
    }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      devShell =
        pkgs.mkShell rec {
          buildInputs = with pkgs; [
            home.defaultPackage.${system}
            go
            gotools
            go-tools
            golangci-lint
            gopls
            go-outline
            gopkgs
            nodejs-18_x
            docker
            docker-credential-helpers
            rsync
            pass
          ];
        };

      defaultPackage =
        with import nixpkgs { inherit system; };
        stdenv.mkDerivation rec {
          name = "${slug}-${version}";

          slug = "defn-cloud";
          version = "0.0.1";

          dontUnpack = true;

          installPhase = "mkdir -p $out";

          propagatedBuildInputs = [ ];

          meta = with lib;
            {
              homepage = "https://defn.sh/${slug}";
              description = "nix golang / tilt integration";
              platforms = platforms.linux;
            };
        };
    });
}