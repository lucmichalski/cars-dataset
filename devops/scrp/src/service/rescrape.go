package main

import (
	"context"
	"crypto/x509"
	"io/ioutil"
	"log"
	"time"

	pb "github.com/dioptre/scrp/src/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	address = "frontend.local:4443"
)

//Rescrape : Go through load balancer, 're-scrape'
func Rescrape(in *pb.ScrapeRequest) {
	FrontendCert, _ := ioutil.ReadFile("./frontend.cert")

	// Create CertPool
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(FrontendCert)

	// Create credentials
	credsClient := credentials.NewClientTLSFromCert(roots, "")

	// Dial with specific Transport (with credentials)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(credsClient))
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}

	defer conn.Close()
	client := pb.NewScraperClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = client.Scrape(ctx, in)
	if err != nil {
		log.Fatalf("could not scrape: %v\n", err)
	}
}
