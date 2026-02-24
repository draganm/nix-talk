package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/draganm/nix-talk/nixdaemon"
	"github.com/nix-community/go-nix/pkg/derivation"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "nix-talk",
		Usage: "communicate with the Nix daemon",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "socket",
				Usage: "path to the Nix daemon socket",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add a .drv file to the Nix store",
				Action: func(c *cli.Context) error {
					drvFile := c.Args().First()
					if drvFile == "" {
						return fmt.Errorf("missing .drv file argument")
					}

					content, err := os.ReadFile(drvFile)
					if err != nil {
						return fmt.Errorf("reading drv file: %w", err)
					}

					drv, err := derivation.ReadDerivation(bytes.NewReader(content))
					if err != nil {
						return fmt.Errorf("parsing derivation: %w", err)
					}

					// Collect all references: input derivation paths + input sources.
					var refs []string
					for path := range drv.InputDerivations {
						refs = append(refs, path)
					}
					refs = append(refs, drv.InputSources...)

					conn, err := nixdaemon.Connect(c.String("socket"))
					if err != nil {
						return err
					}
					defer conn.Close()

					name := filepath(drvFile)
					storePath, err := conn.AddToStore(name, content, refs)
					if err != nil {
						return err
					}

					fmt.Println(storePath)
					return nil
				},
			},
			{
				Name:  "build",
				Usage: "build a derivation",
				Action: func(c *cli.Context) error {
					drvFile := c.Args().First()
					if drvFile == "" {
						return fmt.Errorf("missing .drv file argument")
					}

					content, err := os.ReadFile(drvFile)
					if err != nil {
						return fmt.Errorf("reading drv file: %w", err)
					}

					drv, err := derivation.ReadDerivation(bytes.NewReader(content))
					if err != nil {
						return fmt.Errorf("parsing derivation: %w", err)
					}

					conn, err := nixdaemon.Connect(c.String("socket"))
					if err != nil {
						return err
					}
					defer conn.Close()

					drvPath := drvFile
					if !strings.HasPrefix(drvPath, "/nix/store/") {
						// Add to store first to get the store path.
						var refs []string
						for path := range drv.InputDerivations {
							refs = append(refs, path)
						}
						refs = append(refs, drv.InputSources...)

						name := filepath(drvFile)
						drvPath, err = conn.AddToStore(name, content, refs)
						if err != nil {
							return fmt.Errorf("adding to store: %w", err)
						}
						fmt.Fprintf(os.Stderr, "added to store: %s\n", drvPath)
					}

					result, err := conn.BuildDerivation(drvPath, drv, os.Stderr)
					if err != nil {
						return err
					}

					// 0=Built, 1=Substituted, 2=AlreadyValid are all success.
					if result.Status > 2 {
						return fmt.Errorf("build failed (status %d): %s", result.Status, result.ErrorMsg)
					}

					statusNames := map[uint64]string{0: "built", 1: "substituted", 2: "already valid"}
					fmt.Printf("build %s\n", statusNames[result.Status])
					for name, out := range drv.Outputs {
						fmt.Printf("  %s -> %s\n", name, out.Path)
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// filepath extracts the base filename from a path.
func filepath(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx >= 0 {
		return path[idx+1:]
	}
	return path
}
