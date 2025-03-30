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
