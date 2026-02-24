package nixdaemon

const (
	// Magic bytes for handshake.
	ClientMagic = 0x6e697863
	ServerMagic = 0x6478696f

	// Opcodes.
	OpSetOptions       = 19
	OpAddToStore       = 7
	OpBuildDerivation  = 36

	// Stderr markers.
	StderrLast          = 0x616c7473
	StderrWrite         = 0x64617416
	StderrError         = 0x63787470
	StderrNext          = 0x6f6c6d67
	StderrStartActivity = 0x53545254
	StderrStopActivity  = 0x53544f50
	StderrResult        = 0x52534c54

	// Protocol version 1.37.
	ProtocolVersion = (1 << 8) | 37
)
