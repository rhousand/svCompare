{
  description = "svCompare – Sailboat comparison web application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        # -------------------------------------------------------
        # Frontend: build Vue SPA with Vite
        # -------------------------------------------------------
        frontend = pkgs.buildNpmPackage {
          pname = "svcompare-frontend";
          version = "0.1.0";
          src = ./frontend;

          # Run `nix build .#frontend` — it will fail with the correct hash.
          # Paste that hash here.
          npmDepsHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

          buildPhase = "npm run build";
          installPhase = "cp -r dist $out";
        };

        # -------------------------------------------------------
        # Backend: build Go binary with embedded frontend
        # -------------------------------------------------------
        svcompare = pkgs.buildGoModule {
          pname = "svcompare";
          version = "0.1.0";
          src = ./backend;

          # Run `nix build` — it will fail with the correct hash.
          # Paste that hash here.
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

          CGO_ENABLED = "0";
          ldflags = [ "-s" "-w" ];

          # Copy built frontend into the embed path before go build.
          preBuild = ''
            mkdir -p frontend/dist
            cp -r ${frontend}/* frontend/dist/
          '';
        };

      in
      {
        # -------------------------------------------------------
        # Packages
        # -------------------------------------------------------
        packages = {
          default  = svcompare;
          frontend = frontend;

          # Build a Docker image loadable with: docker load < result
          docker = pkgs.dockerTools.buildLayeredImage {
            name = "svcompare";
            tag  = "latest";
            contents = [ svcompare pkgs.cacert ];
            config = {
              Cmd          = [ "/bin/svcompare" ];
              Env          = [
                "DATABASE_PATH=/data/svcompare.db"
                "PORT=8080"
                "GO_ENV=production"
              ];
              ExposedPorts = { "8080/tcp" = { }; };
              Volumes      = { "/data" = { }; };
              WorkingDir   = "/";
            };
          };
        };

        # -------------------------------------------------------
        # Dev shell
        # -------------------------------------------------------
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            air          # Go hot-reload
            nodejs
            nodePackages.npm
            sqlite
            git
          ];

          shellHook = ''
            export DATABASE_PATH="$PWD/data/svcompare.db"
            export JWT_SECRET="dev-secret-change-in-prod"
            export GO_ENV="development"
            export PORT="8080"
            mkdir -p "$PWD/data"
            echo ""
            echo "  svCompare dev environment"
            echo "  ─────────────────────────────────────────"
            echo "  Backend:  cd backend && air"
            echo "  Frontend: cd frontend && npm install && npm run dev"
            echo "  URLs:     http://localhost:5173  (app)"
            echo "            http://localhost:8080  (API)"
            echo "  Login:    admin / admin"
            echo ""
          '';
        };
      }
    );
}
