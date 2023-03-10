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

// JobBatchInfo information required to create a new job batch
type JobBatchInfo struct {
	// the name of the batch (not unique, a user-friendly name)
	Name string `json:"name"`
	// any relevant notes for the batch (not mandatory)
	Notes string `json:"notes,omitempty"`
	// one or more search labels
	Label []string `json:"label,omitempty"`
	// the universally unique host identifier created by pilot
	HostUUID []string `json:"host_uuid"`
	// the unique key of the function to run
	FxKey string `json:"fx_key"`
	// the version of the function to run
	FxVersion int64 `json:"fx_version"`
}
