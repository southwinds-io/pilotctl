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
