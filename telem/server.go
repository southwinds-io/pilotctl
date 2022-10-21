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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"southwinds.dev/artisan/build"
	"southwinds.dev/artisan/merge"
)

const OtDefaultPort = 4317

// OtServer open telemetry server
type OtServer struct {
	server *grpc.Server
	port   int
}

func NewOtServer(port int) *OtServer {
	if port == 0 {
		port = OtDefaultPort
	}
	s := &OtServer{
		server: grpc.NewServer(),
		port:   OtDefaultPort,
	}
	pmetricotlp.RegisterGRPCServer(s.server, &MetricsHandler{})
	return s
}

func (s *OtServer) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("open telementry server failed to listen: %v", err)
	}
	log.Printf("open telemetry service listening on :%v\n", s.port)
	return s.server.Serve(lis)
}

type MetricsHandler struct {
}

func (h *MetricsHandler) Export(_ context.Context, request pmetricotlp.Request) (pmetricotlp.Response, error) {
	var (
		conn string
		ok   bool
	)
	if conn, ok = connectorName(); !ok {
		return pmetricotlp.NewResponse(), fmt.Errorf("telemetry connector not defined, skipping telemetry recording")
	}
	data, err := request.MarshalJSON()
	if err != nil {
		return pmetricotlp.NewResponse(), fmt.Errorf("cannot marshal request to json: %s", err)
	}
	b64Data := base64.StdEncoding.EncodeToString(data)
	cmd := fmt.Sprintf("%s %s", conn, b64Data)
	var out string
	out, err = build.Exe(cmd, ".", merge.NewEnVarFromSlice(os.Environ()), false)
	if err != nil {
		return pmetricotlp.NewResponse(), fmt.Errorf("internal error executing connector %s: %s", conn, err)
	}
	result := new(connResult)
	err = json.Unmarshal([]byte(out), result)
	if err != nil {
		return pmetricotlp.NewResponse(), fmt.Errorf("connector %s gave an invalid result format, unmarshalling failed: %s", conn, err)
	}
	if len(result.Error) > 0 {
		err = fmt.Errorf("connector %s  returned error: %s", conn, result.Error)
	}
	resp := pmetricotlp.NewResponse()
	if len(result.Error) > 0 {
		resp.PartialSuccess().SetRejectedDataPoints(int64(result.TotalEntries - result.SuccessfulEntries))
		resp.PartialSuccess().SetErrorMessage(result.Error)
	}
	return resp, nil
}

// the name of the connector to use
func connectorName() (string, bool) {
	n := os.Getenv("PILOTCTL_TELEM_CONN")
	return n, len(n) > 0
}

// the result of the connector execution
type connResult struct {
	Error             string `json:"e,omitempty"`
	TotalEntries      int    `json:"t,omitempty"`
	SuccessfulEntries int    `json:"s,omitempty"`
}
