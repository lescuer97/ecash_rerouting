package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"github.com/elnosh/gonuts/wallet"
	"github.com/lescuer97/ecash_rerouting/internal/communication"
	"github.com/lightningnetwork/lnd/lnrpc"
)

func ListenToCustomLightningMessages(ctx context.Context, lnClient lnrpc.LightningClient, wallet *wallet.Wallet) chan bool {

	messageSubRequest := lnrpc.SubscribeCustomMessagesRequest{}
	stream, err := lnClient.SubscribeCustomMessages(ctx, &messageSubRequest)
	if err != nil {
		log.Fatalf("Error setting up lightning comms: %v", err)
	}

	done := make(chan bool)
	log.Println("Running custom message watcher")

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				done <- true //means stream is finished
				return
			}
			if err != nil {
				log.Fatalf("cannot receive %w", err)
			}

			switch resp.Type {
			case communication.REBALANCING_ATTEMPT:
				log.Printf("\n received message: %v \n NICE \n", communication.REBALANCING_ATTEMPT)

				var rebalanceReq communication.ECASH_REBALANCE_REQUEST_REQUEST

				err := json.Unmarshal(resp.Data, &rebalanceReq)
				if err != nil {
					log.Fatalf("json.Unmarshal(resp.Data, &rebalanceReq) %w", err)
				}

				// rebalanceReq.A
				log.Printf("\n rebalancing for: %v msats \n ", rebalanceReq.AmountMsat)

				pubkey := wallet.GetReceivePubkey()

				rebalanceResponse := communication.ECASH_REBALANCE_REQUEST_RESPONSE{
					Pubkey: pubkey.SerializeCompressed(),
					Id:     rebalanceReq.Id,
				}

				bytes, err := json.Marshal(rebalanceResponse)
				if err != nil {
					log.Fatalf("json.Marshal(rebalanceReq) %v", err)
				}

				customRequest := lnrpc.SendCustomMessageRequest{
					Type: communication.REBALANCING_PUBKEY,
					Peer: resp.Peer,
					Data: bytes,
				}

				_, err = lnClient.SendCustomMessage(ctx, &customRequest)

				if err != nil {
					log.Fatalf("cannot send custom message %v", err)
				}

			case communication.REBALANCING_PUBKEY:
				var pubkeyRebalance communication.ECASH_REBALANCE_ATTEMPT_RESPONSE

				err := json.Unmarshal(resp.Data, &pubkeyRebalance)
				if err != nil {
					log.Fatalf("json.Unmarshal(resp.Data, &pubkeyRebalance) %+v", err)
				}

				fmt.Printf("recieved rebalance pubkey: %+v", pubkeyRebalance)

			}

			err = stream.CloseSend()
			if err != nil {
				log.Fatalln("could not close stream", err)
			}

		}
	}()

	// <-done //we will wait until all response is received
	// log.Printf("finished listening to custom messages")

	return done
}
