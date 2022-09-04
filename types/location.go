/*
  pilot control service
  © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package types

// Location host location
type Location struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// Area host area within a Location
type Area struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Org host organisation
type Org struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
