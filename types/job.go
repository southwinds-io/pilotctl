/*
  pilot control service
  © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package types

// Job a representation of a job in the database
type Job struct {
	Id         int64    `json:"id"`
	HostUUID   string   `json:"host_uuid"`
	JobBatchId int64    `json:"job_batch_id"`
	FxKey      string   `json:"fx_key"`
	FxVersion  int64    `json:"fx_version"`
	Created    string   `json:"created"`
	Started    string   `json:"started"`
	Completed  string   `json:"completed"`
	Log        string   `json:"log"`
	Error      bool     `json:"error"`
	OrgGroup   string   `json:"org_group"`
	Org        string   `json:"org"`
	Area       string   `json:"area"`
	Location   string   `json:"location"`
	Tag        []string `json:"tag"`
}
