package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/lescuer97/ecash_rerouting/internal/communication"
	"github.com/lescuer97/ecash_rerouting/internal/wallet"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc/metadata"
)

func IsEnoughLiquidity(chanStatus *lnrpc.Channel, htlc *routerrpc.ForwardHtlcInterceptRequest) bool {


	if (chanStatus.LocalBalance * 1000) < int64(htlc.OutgoingAmountMsat) {
		return false
	}

	return true
}

func GetLiquidityDisparity(chanStatus *lnrpc.Channel, htlc *routerrpc.ForwardHtlcInterceptRequest) uint64 {
	return htlc.OutgoingAmountMsat - uint64(chanStatus.LocalBalance*1000)
}

func ChannelByPeerId(channels *lnrpc.ListChannelsResponse, chanId uint64) *lnrpc.Channel {
	for _, v := range channels.Channels {
		if v.ChanId == chanId {
			return v

		}

	}
	return nil
}

type RebalancingStates struct {
    states map[uuid.UUID]RebalanceAttempt
    sync.Mutex
}

type RebalanceAttempt struct {
    Approved bool
    Amount  uint64
}

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	lightningComms, err := communication.SetUpLightningComms(os.Getenv(communication.LND_HOST_1), os.Getenv(communication.LND_TLS_CERT_1), os.Getenv(communication.LND_MACAROON_1))
	if err != nil {
		log.Fatalf("Error setting up lightning comms: %v", err)
	}

	routerClient := routerrpc.NewRouterClient(lightningComms.LndRpcClient)

	ctx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", lightningComms.Macaroon)
	lnClient := lnrpc.NewLightningClient(lightningComms.LndRpcClient)

	stream, err := routerClient.HtlcInterceptor(ctx)
	if err != nil {
		log.Fatalf("Error setting up lightning comms: %v", err)
	}

	done := make(chan bool)
	fmt.Println("Channel interceptor running")

    wallet, err := wallet.SetUpWallet(".", os.Getenv(wallet.ACTIVE_MINT))
	if err != nil {
		log.Fatalf("wallet.SetUpWallet(os.Getenv(wallet.WALLET_DB_1), os.Getenv(wallet.ACTIVE_MINT)): %v", err)
	}



    

    thread :=   ListenToCustomLightningMessages(ctx,lnClient, wallet)
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				done <- true //means stream is finished
				return
			}
			if err != nil {
				log.Fatalf("cannot receive %v", err)
			}

			listChanRequest := lnrpc.ListChannelsRequest{
				ActiveOnly: true,
			}
			listChannel, err := lnClient.ListChannels(ctx, &listChanRequest)

			if err != nil {
				log.Fatalf("cannot receive %v", err)
			}

			channel := ChannelByPeerId(listChannel, resp.OutgoingRequestedChanId)
			channel.GetPeerScidAlias()

			if channel == nil {
				log.Panic("No available channel")
			}

			// enoughLiquidity := IsEnoughLiquidity(channel, resp)
			//
			// if enoughLiquidity {
			//     fmt.Println("continuing payment")
			// 	response := &routerrpc.ForwardHtlcInterceptResponse{
			// 		IncomingCircuitKey: resp.IncomingCircuitKey,
			// 		Action:             routerrpc.ResolveHoldForwardAction_RESUME,
			// 	}
			//
			// 	err = stream.Send(response)
			// 	if err != nil {
			// 		log.Fatalf("cannot send %v", err)
			// 	}
			//
			// }
			//
			//          _ = GetLiquidityDisparity(channel, resp)
			//

			peerKey, err := hex.DecodeString(channel.RemotePubkey)
			if err != nil {
				log.Fatalf("hex.DecodeString(channel.RemotePubkey) %v", err)
			}

            rebalanceReq := communication.ECASH_REBALANCE_REQUEST_REQUEST {
                AmountMsat: 10,

            }

            bytes, err := json.Marshal(rebalanceReq)
			if err != nil {
				log.Fatalf("json.Marshal(rebalanceReq) %v", err)
			}
            

			customRequest := lnrpc.SendCustomMessageRequest{
				Type: communication.REBALANCING_ATTEMPT,
				Peer: peerKey,
				Data: bytes,
			}

			_, err = lnClient.SendCustomMessage(ctx, &customRequest)

			if err != nil {
				log.Fatalf("cannot send custom message %v", err)
			}

            // time.Sleep()


			fmt.Println("continuing payment")
			response := &routerrpc.ForwardHtlcInterceptResponse{
				IncomingCircuitKey: resp.IncomingCircuitKey,
				Action:             routerrpc.ResolveHoldForwardAction_RESUME,
			}

			err = stream.Send(response)
			if err != nil {
				log.Fatalf("cannot send %v", err)
			}
		}
	}()

	<-done //we will wait until all response is received
    <-thread
	log.Printf("finished")

}
