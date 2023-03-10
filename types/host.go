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

// Host monitoring information
type Host struct {
	Id             int64    `json:"id"`
	HostUUID       string   `json:"host_uuid"`
	HostMacAddress string   `json:"host_mac_address"`
	OrgGroup       string   `json:"org_group"`
	Org            string   `json:"org"`
	Area           string   `json:"area"`
	Location       string   `json:"location"`
	Connected      bool     `json:"connected"`
	LastSeen       int64    `json:"last_seen"`
	Since          int      `json:"since"`
	SinceType      string   `json:"since_type"`
	Label          []string `json:"label"`
	Critical       int      `json:"critical"`
	High           int      `json:"high"`
	Medium         int      `json:"medium"`
	Low            int      `json:"low"`
}
