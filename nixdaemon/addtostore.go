package nixdaemon

import (
	"fmt"

	"github.com/nix-community/go-nix/pkg/wire"
)

// AddToStore adds a file to the Nix store using text hashing (text:sha256).
// It returns the resulting store path.
func (c *Conn) AddToStore(name string, content []byte, references []string) (string, error) {
	if err := wire.WriteUint64(c.w, OpAddToStore); err != nil {
		return "", fmt.Errorf("sending opcode: %w", err)
	}

	if err := wire.WriteString(c.w, name); err != nil {
		return "", fmt.Errorf("sending name: %w", err)
	}

	if err := wire.WriteString(c.w, "text:sha256"); err != nil {
		return "", fmt.Errorf("sending hash type: %w", err)
	}

	if err := writeStringSet(c.w, references); err != nil {
		return "", fmt.Errorf("sending references: %w", err)
	}

	if err := wire.WriteBool(c.w, false); err != nil {
		return "", fmt.Errorf("sending repair flag: %w", err)
	}

	if err := writeFramedSource(c.w, content); err != nil {
		return "", fmt.Errorf("sending content: %w", err)
	}

	if err := c.processStderr(nil); err != nil {
		return "", fmt.Errorf("processing stderr: %w", err)
	}

	storePath, err := readValidPathInfo(c.r)
	if err != nil {
		return "", fmt.Errorf("reading valid path info: %w", err)
	}

	return storePath, nil
}
