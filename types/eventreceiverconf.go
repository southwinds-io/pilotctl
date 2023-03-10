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
	"encoding/json"
	"fmt"
	"os"
)

type EventReceiver struct {
	Name string `json:"name,omitempty"`
	URI  string `json:"uri"`
	// optional credentials if authentication is required
	User string `json:"user,omitempty"`
	Pwd  string `json:"pwd,omitempty"`
}

type EventReceivers struct {
	EventReceivers []EventReceiver `json:"event_receivers"`
}

func NewEventPubConf() *EventReceivers {
	confFile := receiverConfigFile()
	if len(confFile) > 0 {
		bytes, err := os.ReadFile(confFile)
		if err != nil {
			return nil
		}
		var conf EventReceivers
		err = json.Unmarshal(bytes, &conf)
		if err != nil {
			fmt.Printf("ERROR: cannot unmarshal event reciever configuration: %s; event receivers have been disabled\n", err)
			return nil
		}
		return &conf
	}
	return nil
}
