package communication

import (
	"crypto/x509"
	"fmt"
	"log"

	"github.com/elnosh/gonuts/cashu"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	LND_HOST_1     = "LND_GRPC_HOST_1"
	LND_TLS_CERT_1 = "LND_TLS_CERT_1"
	LND_MACAROON_1 = "LND_MACAROON_1"

	LND_HOST_2     = "LND_GRPC_HOST_2"
	LND_TLS_CERT_2 = "LND_TLS_CERT_2"
	LND_MACAROON_2 = "LND_MACAROON_2"
)

const (
	REBALANCING_ATTEMPT = 42069
	REBALANCING_PUBKEY = 42070
)

type ECASH_REBALANCE_REQUEST_REQUEST struct {
	AmountMsat         uint64
    Id uuid.UUID
}

type ECASH_REBALANCE_REQUEST_RESPONSE struct {
	Pubkey []byte
    Id uuid.UUID
}

type ECASH_REBALANCE_REQUEST struct {
	LockedProofs         cashu.Token
	InvoiceRequest string
}

type ECASH_REBALANCE_ATTEMPT_RESPONSE struct {
	Pubkey []byte
	// InvoiceRequest string
}

type LightingComms struct {
	LndRpcClient *grpc.ClientConn
	Macaroon     string
}

func SetUpLightningComms(host string, pem_cert string, macaroon string) (*LightingComms, error) {
	lightningComs := LightingComms{}

	if host == "" {
		return nil, fmt.Errorf("LND_GRPC_HOST not available")
	}

	if pem_cert == "" {
		return nil, fmt.Errorf("LND_CERT_PATH not available")
	}

	certPool := x509.NewCertPool()

	appendOk := certPool.AppendCertsFromPEM([]byte(pem_cert))

	if !appendOk {
		log.Printf("x509.AppendCertsFromPEM(): failed")
		return nil, fmt.Errorf("x509.AppendCertsFromPEM(): failed")
	}

	certFile := credentials.NewClientTLSFromCert(certPool, "")

	tlsDialOption := grpc.WithTransportCredentials(certFile)

	dialOpts := []grpc.DialOption{
		tlsDialOption,
	}

	clientConn, err := grpc.Dial(host, dialOpts...)

	if err != nil {
		return nil, err
	}

	if macaroon == "" {
		return nil, fmt.Errorf("LND_MACAROON_PATH not available")
	}

	lightningComs.LndRpcClient = clientConn
	lightningComs.Macaroon = macaroon

	return &lightningComs, nil
}
