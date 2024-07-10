package communication

import (
	// "crypto/rand"
	// "errors"
	// "fmt"
	// "time"

	// "github.com/lightningnetwork/lnd"
	// "github.com/lightningnetwork/lnd/lnrpc"
	// "github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	// "github.com/lightningnetwork/lnd/lntypes"
	// "github.com/lightningnetwork/lnd/lnwire"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	LND_HOST     = "LND_GRPC_HOST"
	LND_TLS_CERT = "LND_TLS_CERT"
	LND_MACAROON = "LND_MACAROON"
)

type LightingComms struct {
	LndRpcClient *grpc.ClientConn
	Macaroon     string
}

func SetUpLightningComms() (*LightingComms, error) {
	lightningComs := LightingComms{}

	host := os.Getenv(LND_HOST)

	if host == "" {
		return nil, fmt.Errorf("LND_GRPC_HOST not available")
	}

	pem_cert := os.Getenv(LND_TLS_CERT)

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

	macaroon := os.Getenv(LND_MACAROON)

	if macaroon == "" {
		return nil, fmt.Errorf("LND_MACAROON_PATH not available")
	}

	lightningComs.LndRpcClient = clientConn
	lightningComs.Macaroon = macaroon

	return &lightningComs, nil
}
