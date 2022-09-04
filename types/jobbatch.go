/*
  pilot control service
  © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
