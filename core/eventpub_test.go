/*
  pilot control service
  © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"southwinds.dev/pilotctl/types"
	"testing"
)

func TestSaveConf(t *testing.T) {
	conf := &types.EventReceivers{EventReceivers: []types.EventReceiver{
		{
			URI:  "AAA",
			User: "BBB",
			Pwd:  "CCC",
		},
	}}
	bytes, _ := json.Marshal(conf)
	path, _ := filepath.Abs("ev_receive.json")
	os.WriteFile(path, bytes, os.ModePerm)
}
