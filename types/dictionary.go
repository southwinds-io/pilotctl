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

// Dictionary a key value pair list with name and description
type Dictionary struct {
	// Key a natural key used to uniquely identify this dictionary for the purpose of idempotent opeartions
	Key string `json:"key" yaml:"key"`
	// Name a friendly name for the dictionary
	Name string `json:"name" yaml:"name"`
	// Description describe the purpose / content of the dictionary
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Values a map containing key/value pairs that are the content held by the dictionary
	Values map[string]interface{} `json:"values,omitempty" yaml:"values,omitempty"`
	// Tags a list of string based tags used for categorising the dictionary
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}
