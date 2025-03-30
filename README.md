# Yellowstone-gRPC Transaction Converter

## Overview

This package provides a converter for Solana transactions received via the [yellowstone-grpc](https://github.com/rpcpool/yellowstone-grpc) plugin into more readable and developer-friendly Go types. The converter transforms raw gRPC transaction data into structured types that are easier to work with in Go applications.

## Key Features

- Converts low-level gRPC transaction data into intuitive Go structs
- Provides comprehensive transaction details including:
  - Signature and slot information
  - Compute unit consumption
  - Full instruction breakdown
  - Balance changes (both native and token)
  - Execution logs
- Supports nested inner instructions

## Installation

```bash
go get github.com/dzhisl/geyser-monitor
```

## Usage

## Data Structure

The package defines the following main types in `shared/types.go`:

1. **TransactionDetails** - Top-level transaction container
2. **InstructionDetails** - Individual instruction data
3. **InnerInstructionDetails** - Nested instructions
4. **AccountReference** - Account metadata
5. **BalanceChanges** - Native balance changes
6. **TokenBalanceChanges** - Token balance changes (default from geyser proto)

## Contributing

PRs are appreciated! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Submit a pull request with a clear description

### Example

```go
package main

import (
	"context"
	"fmt"
	"log"

	geyserAdapter "github.com/dzhisl/geyser-converter/geyser"
	"github.com/dzhisl/geyser-converter/shared"
	"github.com/dzhisl/geyser-converter/utils"

	"github.com/mr-tron/base58"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
)

const (
	targetAccount = "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4"
	endpointURL   = "http://yourgrpc:10000"
)

func main() {

	grpcAdapter := geyserAdapter.NewGeyserAdapter()

	conn, err := grpcAdapter.CreateGRPCConnection(endpointURL)
	if err != nil {
		log.Fatalf("Connection failed: %s", err)
	}
	defer conn.Close()

	client := pb.NewGeyserClient(conn)
	stream, err := client.Subscribe(context.Background())
	if err != nil {
		log.Fatalf("Failed to create stream: %s", err)
	}

	if err := stream.Send(grpcAdapter.CreateSubscriptionRequest(targetAccount)); err != nil {
		log.Fatalf("Subscription failed: %s", err)
	}

	fmt.Printf("🔭 Monitoring transactions for account: %s\n", targetAccount)

	for {
		update, err := stream.Recv()
		if err != nil {
			log.Fatalf("Stream error: %s", err)
		}
		go processUpdate(update)
	}
}

func processUpdate(update *pb.SubscribeUpdate) *shared.TransactionDetails {
	//ping messages don't have transaction inside
	if update.GetTransaction() == nil {
		return nil
	}

	tx := update.GetTransaction()
	signature := tx.GetTransaction().GetSignature()

	//handle transaction with empty signature
	if signature == nil {
		return nil
	}

	sigStr := base58.Encode(signature)
	txDetails, err := utils.ProcessTransactionToStruct(tx, sigStr)
	if err != nil {
		fmt.Printf("Error processing tx details: %s", err)
	}
	fmt.Println("Proccessed transaction", txDetails.Signature)
	return txDetails
}

```
