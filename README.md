# nix-talk

A Go library and CLI for communicating with the Nix daemon over its binary Unix socket protocol. Supports adding `.drv` files to the Nix store and building derivations.

## Usage

```bash
# Add a .drv file to the store
go run . add /path/to/hello.drv

# Build a derivation (already in the store)
go run . build /nix/store/...-hello.drv

# Build a local .drv file (adds to store first, then builds)
go run . build /path/to/hello.drv

# Use a custom daemon socket
go run . --socket /path/to/socket build /nix/store/...-hello.drv
```

## Creating a test derivation

```bash
nix-instantiate -E 'derivation {
  name = "hello";
  builder = "/bin/sh";
  args = ["-c" "echo hello > $out"];
  system = builtins.currentSystem;
}'
```

## Project structure

```
nixdaemon/           -- library for talking to the Nix daemon
  protocol.go        -- constants (magic bytes, opcodes, stderr markers)
  conn.go            -- connection, handshake, SetOptions
  stderr.go          -- stderr message processing
  wiretypes.go       -- wire serialization helpers
  addtostore.go      -- AddToStore operation
  buildderivation.go -- BuildDerivation operation
main.go              -- CLI (add + build subcommands)
```

## Credits

Built by [Anthropic](https://anthropic.com)'s Claude Code.
