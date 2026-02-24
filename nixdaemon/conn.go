package nixdaemon

import (
	"fmt"
	"io"
	"net"

	"github.com/nix-community/go-nix/pkg/wire"
)

// Conn represents a connection to the Nix daemon.
type Conn struct {
	conn          net.Conn
	r             io.Reader
	w             io.Writer
	DaemonVersion uint64
}

// Connect establishes a connection to the Nix daemon and performs the handshake.
func Connect(socketPath string) (*Conn, error) {
	if socketPath == "" {
		socketPath = "/nix/var/nix/daemon-socket/socket"
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("connecting to nix daemon: %w", err)
	}

	c := &Conn{
		conn: conn,
		r:    conn,
		w:    conn,
	}

	if err := c.handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	return c, nil
}

func (c *Conn) handshake() error {
	// Send client magic.
	if err := wire.WriteUint64(c.w, ClientMagic); err != nil {
		return fmt.Errorf("sending client magic: %w", err)
	}

	// Read server magic.
	serverMagic, err := wire.ReadUint64(c.r)
	if err != nil {
		return fmt.Errorf("reading server magic: %w", err)
	}
	if serverMagic != ServerMagic {
		return fmt.Errorf("unexpected server magic: 0x%x", serverMagic)
	}

	// Read daemon version.
	c.DaemonVersion, err = wire.ReadUint64(c.r)
	if err != nil {
		return fmt.Errorf("reading daemon version: %w", err)
	}

	// Send client version.
	if err := wire.WriteUint64(c.w, ProtocolVersion); err != nil {
		return fmt.Errorf("sending client version: %w", err)
	}

	// Send CPU affinity (0 = don't set).
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return fmt.Errorf("sending cpu affinity: %w", err)
	}

	// Send reserved field.
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return fmt.Errorf("sending reserved: %w", err)
	}

	// Read daemon version string (unused).
	if _, err := wire.ReadString(c.r, 64*1024); err != nil {
		return fmt.Errorf("reading daemon version string: %w", err)
	}

	// Read trust status.
	trustStatus, err := wire.ReadUint64(c.r)
	if err != nil {
		return fmt.Errorf("reading trust status: %w", err)
	}
	if trustStatus == 2 {
		return fmt.Errorf("daemon reports untrusted client")
	}

	// Process initial handshake stderr.
	if err := c.processStderr(nil); err != nil {
		return fmt.Errorf("handshake stderr: %w", err)
	}

	// Send SetOptions.
	if err := c.setOptions(); err != nil {
		return err
	}

	return nil
}

func (c *Conn) setOptions() error {
	if err := wire.WriteUint64(c.w, OpSetOptions); err != nil {
		return err
	}
	// keepFailed
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return err
	}
	// keepGoing
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return err
	}
	// tryFallback
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return err
	}
	// verbosity
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return err
	}
	// maxBuildJobs
	if err := wire.WriteUint64(c.w, 1); err != nil {
		return err
	}
	// maxSilentTime
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return err
	}
	// useBuildHook
	if err := wire.WriteUint64(c.w, 1); err != nil {
		return err
	}
	// verbosity (again for buildVerbosity)
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return err
	}
	// logType
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return err
	}
	// printBuildTrace
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return err
	}
	// buildCores
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return err
	}
	// useSubstitutes
	if err := wire.WriteUint64(c.w, 1); err != nil {
		return err
	}
	// overrides count
	if err := wire.WriteUint64(c.w, 0); err != nil {
		return err
	}

	return c.processStderr(nil)
}

// Close closes the connection to the daemon.
func (c *Conn) Close() error {
	return c.conn.Close()
}
