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

import "bytes"

// RegistrationRequest information sent by pilot upon host registration
type RegistrationRequest struct {
	Hostname    string   `json:"hostname"`
	HostIP      string   `json:"host_ip"`
	MachineId   string   `json:"machine_id"`
	OS          string   `json:"os"`
	Platform    string   `json:"platform"`
	Virtual     bool     `json:"virtual"`
	TotalMemory float64  `json:"total_memory"`
	CPUs        int      `json:"cpus"`
	MacAddress  []string `json:"mac_address"`
}

// Reader Get a JSON bytes reader for the Serializable
func (r *RegistrationRequest) Reader() (*bytes.Reader, error) {
	jsonBytes, err := r.Bytes()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(*jsonBytes), err
}

// Bytes Get a []byte representing the Serializable
func (r *RegistrationRequest) Bytes() (*[]byte, error) {
	b, err := ToJson(r)
	return &b, err
}

// RegistrationResponse data returned to pilot upon registration
type RegistrationResponse struct {
	// the status of the registration - I: created, U: updated, N: already exist
	Operation string `json:"operation"`
}
