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
