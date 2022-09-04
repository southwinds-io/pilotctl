/*
  pilot control service
  © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package types

// Admission an admission request
type Admission struct {
	HostUUID string   `json:"host_uuid"`
	OrgGroup string   `json:"org_group"`
	Org      string   `json:"org"`
	Area     string   `json:"area"`
	Location string   `json:"location"`
	Label    []string `json:"label"`
}

type Registration struct {
	MacAddress string   `json:"mac_address"`
	OrgGroup   string   `json:"org_group"`
	Org        string   `json:"org"`
	Area       string   `json:"area"`
	Location   string   `json:"location"`
	Label      []string `json:"label"`
}
