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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"testing"
)

// drop .pilot_verify.pgp and .pilot_sign.pgp in the user home
// for the test to read them
func TestPingSignVerify(t *testing.T) {
	resp, err := NewPingResponse(CmdInfo{
		JobId:         1,
		Package:       "test/pack:latest",
		Function:      "run",
		User:          "demouser",
		Pwd:           "asdf1234",
		Verbose:       false,
		Containerised: false,
		Input:         nil,
	}, 15000)
	if err != nil {
		t.Fatal(err)
	}
	puKeyPath, err := KeyFilePath("verify")
	if err != nil {
		t.Fatal(err)
	}
	pubKey, err := ioutil.ReadFile(puKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	err = verify(resp.Envelope, resp.Signature, pubKey)
	if err != nil {
		t.Fatal(err)
	}
}

func verify(obj interface{}, signature string, pubKey []byte) error {
	// decode the  signature
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("verify => cannot decode signature string '%s': %s\n", signature, err)
	}
	// obtain the object checksum
	sum, err := checksum(obj)
	if err != nil {
		return fmt.Errorf("verify => cannot calculate checksum: %s\n", err)
	}
	// load verification key from activation key
	pgp, err := LoadPGPBytes(pubKey)
	if err != nil {
		return fmt.Errorf("verify => cannot load host verification key: %s", err)
	}
	// check loaded key is not private
	if pgp.HasPrivate() {
		return fmt.Errorf("verify => verification key should be public, private key found\n")
	}
	// verify digital signature
	return pgp.Verify(sum, sig)
}

func TestHwId(t *testing.T) {
	// r, _ := regexp.Compile(".*Hardware UUID: (?P<HW_ID>.*)\\b")
	// idBytes := r.Find(out)
	if len(getHwId()) == 0 {
		t.FailNow()
	}
}
