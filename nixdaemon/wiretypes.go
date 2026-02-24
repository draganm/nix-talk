package nixdaemon

import (
	"fmt"
	"io"
	"sort"

	"github.com/nix-community/go-nix/pkg/derivation"
	"github.com/nix-community/go-nix/pkg/wire"
)

// BuildResult holds the result of a build operation.
type BuildResult struct {
	Status    uint64
	ErrorMsg  string
	TimesBuilt uint64
	IsNonDeterministic bool
	StartTime uint64
	StopTime  uint64
}

func writeStringSet(w io.Writer, paths []string) error {
	sorted := make([]string, len(paths))
	copy(sorted, paths)
	sort.Strings(sorted)

	if err := wire.WriteUint64(w, uint64(len(sorted))); err != nil {
		return err
	}
	for _, p := range sorted {
		if err := wire.WriteString(w, p); err != nil {
			return err
		}
	}
	return nil
}

func readStringSet(r io.Reader) ([]string, error) {
	count, err := wire.ReadUint64(r)
	if err != nil {
		return nil, err
	}
	result := make([]string, count)
	for i := range result {
		s, err := wire.ReadString(r, 64*1024)
		if err != nil {
			return nil, err
		}
		result[i] = s
	}
	return result, nil
}

func writeDerivationWire(w io.Writer, drv *derivation.Derivation) error {
	// Outputs: sorted by name.
	outputNames := make([]string, 0, len(drv.Outputs))
	for name := range drv.Outputs {
		outputNames = append(outputNames, name)
	}
	sort.Strings(outputNames)

	if err := wire.WriteUint64(w, uint64(len(outputNames))); err != nil {
		return err
	}
	for _, name := range outputNames {
		out := drv.Outputs[name]
		if err := wire.WriteString(w, name); err != nil {
			return err
		}
		if err := wire.WriteString(w, out.Path); err != nil {
			return err
		}
		if err := wire.WriteString(w, out.HashAlgorithm); err != nil {
			return err
		}
		if err := wire.WriteString(w, out.Hash); err != nil {
			return err
		}
	}

	// Input sources (sorted).
	if err := writeStringSet(w, drv.InputSources); err != nil {
		return err
	}

	// Platform.
	if err := wire.WriteString(w, drv.Platform); err != nil {
		return err
	}

	// Builder.
	if err := wire.WriteString(w, drv.Builder); err != nil {
		return err
	}

	// Arguments.
	if err := wire.WriteUint64(w, uint64(len(drv.Arguments))); err != nil {
		return err
	}
	for _, arg := range drv.Arguments {
		if err := wire.WriteString(w, arg); err != nil {
			return err
		}
	}

	// Env: sorted by key.
	envKeys := make([]string, 0, len(drv.Env))
	for k := range drv.Env {
		envKeys = append(envKeys, k)
	}
	sort.Strings(envKeys)

	if err := wire.WriteUint64(w, uint64(len(envKeys))); err != nil {
		return err
	}
	for _, k := range envKeys {
		if err := wire.WriteString(w, k); err != nil {
			return err
		}
		if err := wire.WriteString(w, drv.Env[k]); err != nil {
			return err
		}
	}

	return nil
}

func readBuildResult(r io.Reader, daemonVersion uint64) (*BuildResult, error) {
	status, err := wire.ReadUint64(r)
	if err != nil {
		return nil, fmt.Errorf("reading build status: %w", err)
	}

	errorMsg, err := wire.ReadString(r, 64*1024)
	if err != nil {
		return nil, fmt.Errorf("reading error message: %w", err)
	}

	result := &BuildResult{
		Status:   status,
		ErrorMsg: errorMsg,
	}

	if daemonVersion >= (1<<8)|29 {
		result.TimesBuilt, err = wire.ReadUint64(r)
		if err != nil {
			return nil, err
		}
		isNonDet, err := wire.ReadUint64(r)
		if err != nil {
			return nil, err
		}
		result.IsNonDeterministic = isNonDet != 0
		result.StartTime, err = wire.ReadUint64(r)
		if err != nil {
			return nil, err
		}
		result.StopTime, err = wire.ReadUint64(r)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func readValidPathInfo(r io.Reader) (string, error) {
	// Read store path.
	storePath, err := wire.ReadString(r, 64*1024)
	if err != nil {
		return "", fmt.Errorf("reading store path: %w", err)
	}

	// Deriver.
	if _, err := wire.ReadString(r, 64*1024); err != nil {
		return "", err
	}

	// NAR hash.
	if _, err := wire.ReadString(r, 64*1024); err != nil {
		return "", err
	}

	// References.
	if _, err := readStringSet(r); err != nil {
		return "", err
	}

	// Registration time.
	if _, err := wire.ReadUint64(r); err != nil {
		return "", err
	}

	// NAR size.
	if _, err := wire.ReadUint64(r); err != nil {
		return "", err
	}

	// Ultimate.
	if _, err := wire.ReadUint64(r); err != nil {
		return "", err
	}

	// Sigs.
	if _, err := readStringSet(r); err != nil {
		return "", err
	}

	// CA.
	if _, err := wire.ReadString(r, 64*1024); err != nil {
		return "", err
	}

	return storePath, nil
}

func writeFramedSource(w io.Writer, data []byte) error {
	if err := wire.WriteUint64(w, uint64(len(data))); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	// Terminator.
	if err := wire.WriteUint64(w, 0); err != nil {
		return err
	}
	return nil
}
