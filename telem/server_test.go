/*
pilot control service
© 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
Contributors to this project, hereby assign copyright in this code to the project,
to be licensed under the same terms as the rest of the code.
*/
package telem

import (
	"context"
	"fmt"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"testing"
	"time"
)

// TestLaunchServer launches a local OT Gateway
func TestLaunchServer(t *testing.T) {
	server := NewOtServer(0)
	server.Start()
}

// TestClient exports metrics to the gateway
func TestClient(t *testing.T) {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", OtDefaultPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pmetricotlp.NewGRPCClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// get metrics
	data, err := os.ReadFile("metrics.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
	req := pmetricotlp.NewRequest()
	err = req.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf(err.Error())
	}
	resp, err := c.Export(ctx, req)
	if err != nil {
		log.Fatalf("could not export metrics: %v", err)
	}
	fmt.Printf("%v", resp)
}
