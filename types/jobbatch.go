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

import "time"

// JobBatch a representation of a batch in the database
type JobBatch struct {
	// the id of the job batch
	BatchId int64 `json:"batch_id"`
	// the name of the batch (not unique, a user-friendly name)
	Name string `json:"name"`
	// any relevant notes for the batch (not mandatory)
	Notes string `json:"notes,omitempty"`
	// creation time
	Created time.Time `json:"created"`
	// one or more search labels
	Label []string `json:"label,omitempty"`
	// owner
	Owner string `json:"owner"`
	// jobs
	Jobs int `json:"jobs"`
}
