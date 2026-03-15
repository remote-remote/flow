{ pkgs, lib, config, inputs, ... }:

let
  pkgs-unstable = import inputs.nixpkgs-unstable { system = pkgs.stdenv.system; };
in
{
  languages.go = {
    enable = true;
    package = pkgs-unstable.go; # Use Go from nixpkgs-unstable (1.24)
  };
  packages = [ 
    pkgs.oha
  ];
}
