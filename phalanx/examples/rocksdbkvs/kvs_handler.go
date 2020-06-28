package rocksdbkvs

import (
	"log"

	"github.com/getumen/doctrine/phalanx"
	"github.com/getumen/doctrine/phalanx/phalanxpb"
)

type commandHandler struct{}

func (c *commandHandler) Apply(
	regionName string,
	command *phalanxpb.Command,
	stableStorage phalanx.StableStore,
) {
	switch command.Command {
	case "PUT":
		batch := stableStorage.CreateBatch()
		for i := range command.KeyValues {
			batch.Put(
				regionName,
				command.KeyValues[i].Key,
				command.KeyValues[i].Value,
			)
		}
		stableStorage.Write(batch)
	default:
		log.Fatalf("undefined command %s", command.Command)
	}
}
