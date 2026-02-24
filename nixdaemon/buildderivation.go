package nixdaemon

import (
	"fmt"
	"io"

	"github.com/nix-community/go-nix/pkg/derivation"
	"github.com/nix-community/go-nix/pkg/wire"
)

// BuildDerivation sends a build request to the Nix daemon.
// Build logs are written to logWriter (can be nil).
func (c *Conn) BuildDerivation(drvPath string, drv *derivation.Derivation, logWriter io.Writer) (*BuildResult, error) {
	if err := wire.WriteUint64(c.w, OpBuildDerivation); err != nil {
		return nil, fmt.Errorf("sending opcode: %w", err)
	}

	if err := wire.WriteString(c.w, drvPath); err != nil {
		return nil, fmt.Errorf("sending drv path: %w", err)
	}

	if err := writeDerivationWire(c.w, drv); err != nil {
		return nil, fmt.Errorf("sending derivation: %w", err)
	}

	// buildMode = 0 (normal)
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return nil, fmt.Errorf("sending build mode: %w", err)
	}

	if err := c.processStderr(logWriter); err != nil {
		return nil, fmt.Errorf("processing stderr: %w", err)
	}

	result, err := readBuildResult(c.r, c.DaemonVersion)
	if err != nil {
		return nil, fmt.Errorf("reading build result: %w", err)
	}

	return result, nil
}
