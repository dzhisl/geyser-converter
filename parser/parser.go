package parser

import (
	"github.com/dzhisl/geyser-converter/shared"
	"github.com/mr-tron/base58"
	"github.com/near/borsh-go"
)

// example of parsing anchor self CPI log for pump.fun
func parseSelfCpiLog(txDetails *shared.TransactionDetails) (*PumpFunAnchorSelfCPILogSwapAction, error) {
	insts := txDetails.Instructions

	for _, inst := range insts {
		for _, innerInst := range inst.InnerInstructions {
			if len(innerInst.Accounts) == 1 &&
				innerInst.Accounts[0].PublicKey == "Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1" &&
				innerInst.ProgramID.PublicKey == "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P" {
				d, err := AnchorSelfCPILogSwapParser([]byte(innerInst.Data))
				if err != nil {
					return nil, err
				}
				return d, nil
			}

		}
	}
	return nil, nil
}

type PumpFunAnchorSelfCPILogSwapAction struct {
	Mint                 string `json:"mint"`
	SolAmount            uint64 `json:"solAmount"`
	TokenAmount          uint64 `json:"tokenAmount"`
	IsBuy                bool   `json:"isBuy"`
	User                 string `json:"user"`
	Timestamp            int64  `json:"timestamp"`
	VirtualSolReserves   uint64 `json:"virtualSolReserves"`
	VirtualTokenReserves uint64 `json:"virtualTokenReserves"`
}

func AnchorSelfCPILogSwapParser(decodedData []byte) (*PumpFunAnchorSelfCPILogSwapAction, error) {
	var data AnchorSelfCPILogSwapData
	err := borsh.Deserialize(&data, decodedData)
	if err != nil {
		return nil, err
	}

	action := PumpFunAnchorSelfCPILogSwapAction{

		Mint:                 base58.Encode(data.Mint[:]),
		SolAmount:            data.SolAmount,
		TokenAmount:          data.TokenAmount,
		IsBuy:                data.IsBuy,
		User:                 base58.Encode(data.User[:]),
		Timestamp:            data.Timestamp,
		VirtualTokenReserves: data.VirtualTokenReserves,
		VirtualSolReserves:   data.VirtualSolReserves,
	}

	return &action, nil
}

type AnchorSelfCPILogSwapData struct {
	Discriminator        [16]byte
	Mint                 [32]byte
	SolAmount            uint64
	TokenAmount          uint64
	IsBuy                bool
	User                 [32]byte
	Timestamp            int64
	VirtualSolReserves   uint64
	VirtualTokenReserves uint64
}
