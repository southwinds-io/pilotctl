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

import "southwinds.dev/artisan/data"

// Cmd command information for remote host execution
type Cmd struct {
	// the natural key uniquely identifying the command
	Key string `json:"key"`
	// description of the command
	Description string `json:"description"`
	// the package to use
	Package string `json:"package"`
	// the function in the package to call
	Function string `json:"function"`
	// the function input information
	Input *data.Input `json:"input"`
	// the package registry user
	User string `json:"user"`
	// the package registry password
	Pwd string `json:"pwd"`
	// enables verbose output
	Verbose bool `json:"verbose"`
	// run command in runtime
	Containerised bool `json:"containerised"`
}
