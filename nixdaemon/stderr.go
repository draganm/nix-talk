package nixdaemon

import (
	"fmt"
	"io"

	"github.com/nix-community/go-nix/pkg/wire"
)

func (c *Conn) processStderr(logWriter io.Writer) error {
	for {
		marker, err := wire.ReadUint64(c.r)
		if err != nil {
			return fmt.Errorf("reading stderr marker: %w", err)
		}

		switch marker {
		case StderrLast:
			return nil

		case StderrWrite:
			msg, err := wire.ReadString(c.r, 64*1024)
			if err != nil {
				return fmt.Errorf("reading stderr write: %w", err)
			}
			if logWriter != nil {
				logWriter.Write([]byte(msg))
			}

		case StderrNext:
			msg, err := wire.ReadString(c.r, 64*1024)
			if err != nil {
				return fmt.Errorf("reading stderr next: %w", err)
			}
			if logWriter != nil {
				logWriter.Write([]byte(msg))
			}

		case StderrError:
			if c.DaemonVersion >= (1<<8)|26 {
				// Protocol >= 1.26: full error object.
				if _, err := wire.ReadString(c.r, 64*1024); err != nil { // type (always "Error")
					return fmt.Errorf("reading error type: %w", err)
				}
				if _, err := wire.ReadUint64(c.r); err != nil { // level
					return fmt.Errorf("reading error level: %w", err)
				}
				if _, err := wire.ReadString(c.r, 64*1024); err != nil { // name
					return fmt.Errorf("reading error name: %w", err)
				}
				msg, err := wire.ReadString(c.r, 256*1024) // msg
				if err != nil {
					return fmt.Errorf("reading error msg: %w", err)
				}
				if _, err := wire.ReadUint64(c.r); err != nil { // havePos (always 0)
					return fmt.Errorf("reading havePos: %w", err)
				}
				nrTraces, err := wire.ReadUint64(c.r)
				if err != nil {
					return fmt.Errorf("reading nrTraces: %w", err)
				}
				for i := uint64(0); i < nrTraces; i++ {
					if _, err := wire.ReadUint64(c.r); err != nil { // havePos
						return err
					}
					if _, err := wire.ReadString(c.r, 256*1024); err != nil { // hint
						return err
					}
				}
				return fmt.Errorf("nix daemon error: %s", msg)
			} else {
				// Protocol < 1.26: legacy format.
				errMsg, err := wire.ReadString(c.r, 64*1024)
				if err != nil {
					return fmt.Errorf("reading error msg: %w", err)
				}
				if _, err := wire.ReadUint64(c.r); err != nil { // exit status
					return fmt.Errorf("reading exit status: %w", err)
				}
				return fmt.Errorf("nix daemon error: %s", errMsg)
			}

		case StderrStartActivity:
			if _, err := wire.ReadUint64(c.r); err != nil { // id
				return err
			}
			if _, err := wire.ReadUint64(c.r); err != nil { // verbosity
				return err
			}
			if _, err := wire.ReadUint64(c.r); err != nil { // type
				return err
			}
			if _, err := wire.ReadString(c.r, 64*1024); err != nil { // msg
				return err
			}
			if err := c.readFields(); err != nil { // fields
				return err
			}
			if _, err := wire.ReadUint64(c.r); err != nil { // parent
				return err
			}

		case StderrStopActivity:
			if _, err := wire.ReadUint64(c.r); err != nil { // id
				return err
			}

		case StderrResult:
			if _, err := wire.ReadUint64(c.r); err != nil { // id
				return err
			}
			if _, err := wire.ReadUint64(c.r); err != nil { // type
				return err
			}
			if err := c.readFields(); err != nil { // fields
				return err
			}

		default:
			return fmt.Errorf("unknown stderr marker: 0x%x", marker)
		}
	}
}

func (c *Conn) readFields() error {
	count, err := wire.ReadUint64(c.r)
	if err != nil {
		return err
	}
	for i := uint64(0); i < count; i++ {
		tag, err := wire.ReadUint64(c.r)
		if err != nil {
			return err
		}
		switch tag {
		case 0: // uint64
			if _, err := wire.ReadUint64(c.r); err != nil {
				return err
			}
		case 1: // string
			if _, err := wire.ReadString(c.r, 64*1024); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown field tag: %d", tag)
		}
	}
	return nil
}
