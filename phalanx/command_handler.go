package phalanx

import "github.com/getumen/doctrine/phalanx/phalanxpb"

// CommandHandler provides command hadler
type CommandHandler interface {
	// Apply applies the command to the stableStorage
	Apply(
		regioin string,
		command *phalanxpb.Command,
		stableStorage StableStore,
	)
}
