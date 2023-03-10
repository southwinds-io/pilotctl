/*
   Pilot Control Service
   Copyright (C) 2022-Present SouthWinds Tech Ltd - www.southwinds.io

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package types

import (
	"bytes"
	"fmt"
	"time"
)

// NewPingResponse creates a new ping response
func NewPingResponse(cmdInfo CmdInfo, pingInterval time.Duration) (*PingResponse, error) {
	// create a signature for the envelope
	envelope := PingResponseEnvelope{
		Command:  cmdInfo,
		Interval: pingInterval,
	}
	signature, err := sign(envelope)
	if err != nil {
		return nil, fmt.Errorf("cannot sign ping response: %s", err)
	}
	return &PingResponse{
		Signature: signature,
		Envelope:  envelope,
	}, nil
}

// PingResponse a command for execution with a job reference
type PingResponse struct {
	// the envelope signature
	Signature string `json:"signature"`
	// the signed content sent to pilot
	Envelope PingResponseEnvelope `json:"envelope"`
}

// PingResponseEnvelope contains the signed content sent to pilot
type PingResponseEnvelope struct {
	// the information about the command to execute
	Command CmdInfo `json:"value"`
	// the ping interval
	Interval time.Duration `json:"interval"`
}

type PingRequest struct {
	Result *JobResult `json:"result,omitempty"`
}

func (r *PingRequest) Reader() (*bytes.Reader, error) {
	jsonBytes, err := r.Bytes()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(*jsonBytes), err
}

func (r *PingRequest) Bytes() (*[]byte, error) {
	b, err := ToJson(r)
	return &b, err
}
