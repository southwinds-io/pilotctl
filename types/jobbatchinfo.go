/*
  pilot control service
  © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
