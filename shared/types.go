package shared

import pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"

type TransactionDetails struct {
	Signature            string
	ComputeUnitsConsumed uint64
	Slot                 uint64
	Instructions         []InstructionDetails
	BalanceChanges       []BalanceChanges
	TokenBalanceChanges  TokenBalanceChanges
	Logs                 []string
}

type TokenBalanceChanges struct {
	PreTokenBalances  []*pb.TokenBalance
	PostTokenBalances []*pb.TokenBalance
}

type BalanceChanges struct {
	Account       AccountReference
	BalanceBefore uint64
	BalanceAfter  uint64
}

type InstructionDetails struct {
	Index             int
	ProgramID         AccountReference
	Data              []byte
	Accounts          []AccountReference
	InnerInstructions []InnerInstructionDetails
}

type InnerInstructionDetails struct {
	OuterIndex int
	InnerIndex int
	ProgramID  AccountReference
	Data       []byte
	Accounts   []AccountReference
}

type AccountReference struct {
	PublicKey  string
	IsWritable bool
	IsReadable bool
	IsSigner   bool
	IsLut      bool
}
