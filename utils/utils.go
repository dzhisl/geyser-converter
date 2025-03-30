package utils

import (
	"bytes"
	"errors"

	"github.com/dzhisl/geyser-converter/shared"
	"github.com/mr-tron/base58"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
)

func ProcessTransactionToStruct(tx *pb.SubscribeUpdateTransaction, signature string) (*shared.TransactionDetails, error) {
	if tx == nil || tx.Transaction == nil {
		return nil, errors.New("transaction is nil")
	}

	meta := tx.Transaction.GetMeta()
	msg := tx.GetTransaction().GetTransaction().GetMessage()
	if msg == nil {
		return nil, errors.New("transaction message is nil")
	}

	// Create transaction details structure with basic info
	txDetails := &shared.TransactionDetails{
		Signature:            signature,
		ComputeUnitsConsumed: meta.GetComputeUnitsConsumed(),
		Slot:                 tx.Slot,
		Logs:                 meta.GetLogMessages(),
		TokenBalanceChanges: shared.TokenBalanceChanges{
			PreTokenBalances:  meta.PreTokenBalances,
			PostTokenBalances: meta.PostTokenBalances,
		},
	}

	// Process accounts and build account reference map for fast lookups
	accountKeys := msg.GetAccountKeys()
	signersIndexes := msg.Header.GetNumRequiredSignatures()
	readSigners := msg.Header.GetNumReadonlySignedAccounts()
	readNonSigners := msg.Header.GetNumReadonlyUnsignedAccounts()

	// Create a map for quick access to account properties
	accountMap := make(map[string]shared.AccountReference, len(accountKeys))
	accountList := make([]shared.AccountReference, len(accountKeys))

	for i, account := range accountKeys {
		pubKey := base58.Encode(account)
		ref := shared.AccountReference{
			PublicKey: pubKey,
		}

		switch {
		case i < int(signersIndexes):
			// Writable signer accounts
			ref.IsWritable = true
			ref.IsSigner = true
		case i < int(signersIndexes+readSigners):
			// Readonly signer accounts
			ref.IsReadable = true
			ref.IsSigner = true
		case i < int(int(signersIndexes)+int(readSigners)+(len(accountKeys)-int(signersIndexes+readSigners+readNonSigners))):
			// Writable non-signer accounts
			ref.IsWritable = true
		default:
			// Readonly non-signer accounts
			ref.IsReadable = true
		}

		accountList[i] = ref
		accountMap[pubKey] = ref
	}

	// Process balance changes
	preBalances := meta.GetPreBalances()
	postBalances := meta.GetPostBalances()
	balanceChanges := make([]shared.BalanceChanges, len(accountKeys))

	loadedWritable := meta.GetLoadedWritableAddresses()
	loadedReadonly := meta.GetLoadedReadonlyAddresses()

	for i, account := range accountKeys {
		pubKey := base58.Encode(account)
		accRef := accountMap[pubKey]

		// Update writable/readable status based on loaded addresses
		if contains(loadedWritable, account) {
			accRef.IsWritable = true
			accRef.IsLut = true
		}
		if contains(loadedReadonly, account) {
			accRef.IsReadable = true
			accRef.IsLut = true
		}

		balanceChanges[i] = shared.BalanceChanges{
			Account:       accRef,
			BalanceBefore: preBalances[i],
			BalanceAfter:  postBalances[i],
		}
	}
	txDetails.BalanceChanges = balanceChanges

	// Create merged accounts list for instruction processing
	// Pre-allocate slice with exact capacity needed
	mergedAccounts := make([][]byte, 0, len(accountKeys)+
		len(loadedWritable)+
		len(loadedReadonly))
	mergedAccounts = append(mergedAccounts, accountKeys...)
	mergedAccounts = append(mergedAccounts, loadedWritable...)
	mergedAccounts = append(mergedAccounts, loadedReadonly...)

	// Process instructions
	instructions := msg.GetInstructions()
	txDetails.Instructions = make([]shared.InstructionDetails, len(instructions))

	for idx, inst := range instructions {
		programIdIndex := inst.GetProgramIdIndex()
		if int(programIdIndex) >= len(mergedAccounts) {
			// zap.L().Fatal("Invalid program ID index",
			// 	zap.Uint32("program_id_index", programIdIndex),
			// 	zap.Int("accounts_length", len(mergedAccounts)))
			continue
		}

		programID := mergedAccounts[programIdIndex]
		programIDStr := base58.Encode(programID)

		instruction := shared.InstructionDetails{
			Index: idx + 1,
			ProgramID: shared.AccountReference{
				PublicKey:  programIDStr,
				IsWritable: contains(loadedWritable, programID) || accountMap[programIDStr].IsWritable,
				IsReadable: contains(loadedReadonly, programID) || accountMap[programIDStr].IsReadable,
				IsSigner:   accountMap[programIDStr].IsSigner,
				IsLut:      contains(loadedWritable, programID) || contains(loadedReadonly, programID),
			},
			Data: inst.GetData(),
		}

		// Process accounts
		accounts := inst.GetAccounts()
		instruction.Accounts = make([]shared.AccountReference, len(accounts))

		for accIdx, accIndex := range accounts {
			if int(accIndex) >= len(mergedAccounts) {
				// zap.L().Fatal("Invalid account index",
				// zap.Int("account_index", int(accIndex)),
				// zap.Int("accounts_length", len(mergedAccounts)))
				continue
			}

			accBytes := mergedAccounts[accIndex]
			accStr := base58.Encode(accBytes)

			instruction.Accounts[accIdx] = shared.AccountReference{
				PublicKey:  accStr,
				IsWritable: contains(loadedWritable, accBytes) || accountMap[accStr].IsWritable,
				IsReadable: contains(loadedReadonly, accBytes) || accountMap[accStr].IsReadable,
				IsSigner:   accountMap[accStr].IsSigner,
				IsLut:      contains(loadedWritable, accBytes) || contains(loadedReadonly, accBytes),
			}
		}

		// Process inner instructions
		if innerInsts := getInnerInstructions(meta, uint32(idx)); len(innerInsts) > 0 {
			instruction.InnerInstructions = make([]shared.InnerInstructionDetails, len(innerInsts))

			for innerIdx, inner := range innerInsts {
				innerProgramIdIndex := inner.GetProgramIdIndex()
				if int(innerProgramIdIndex) >= len(mergedAccounts) {
					// zap.L().Fatal("Invalid inner program ID index",
					// 	zap.Uint32("program_id_index", innerProgramIdIndex),
					// 	zap.Int("accounts_length", len(mergedAccounts)))
					continue
				}

				innerInstruction := shared.InnerInstructionDetails{
					OuterIndex: idx + 1,
					InnerIndex: innerIdx + 1,
					ProgramID: shared.AccountReference{
						PublicKey:  base58.Encode(mergedAccounts[innerProgramIdIndex]),
						IsWritable: contains(loadedWritable, programID) || accountMap[programIDStr].IsWritable,
						IsReadable: contains(loadedReadonly, programID) || accountMap[programIDStr].IsReadable,
						IsSigner:   accountMap[programIDStr].IsSigner,
					},
					Data: inner.GetData(),
				}

				// Process inner accounts
				innerAccounts := inner.GetAccounts()
				innerInstruction.Accounts = make([]shared.AccountReference, len(innerAccounts))

				for accIdx, accIndex := range innerAccounts {
					if int(accIndex) >= len(mergedAccounts) {
						// zap.L().Fatal("Invalid inner account index",
						// 	zap.Int("account_index", int(accIndex)),
						// 	zap.Int("accounts_length", len(mergedAccounts)))
						continue
					}

					accBytes := mergedAccounts[accIndex]
					accStr := base58.Encode(accBytes)

					innerInstruction.Accounts[accIdx] = shared.AccountReference{
						PublicKey:  accStr,
						IsWritable: contains(loadedWritable, accBytes) || accountMap[accStr].IsWritable,
						IsReadable: contains(loadedReadonly, accBytes) || accountMap[accStr].IsReadable,
						IsSigner:   accountMap[accStr].IsSigner,
					}
				}

				instruction.InnerInstructions[innerIdx] = innerInstruction
			}
		}

		txDetails.Instructions[idx] = instruction
	}

	return txDetails, nil
}

// Simple contains function using bytes.Equal
func contains(s [][]byte, e []byte) bool {
	for _, a := range s {
		if bytes.Equal(a, e) {
			return true
		}
	}
	return false
}

// getInnerInstructions retrieves inner instructions for a given index
func getInnerInstructions(meta *pb.TransactionStatusMeta, index uint32) []*pb.InnerInstruction {
	if meta == nil {
		return nil
	}
	for _, inner := range meta.GetInnerInstructions() {
		if inner.GetIndex() == index {
			return inner.GetInstructions()
		}
	}
	return nil
}
