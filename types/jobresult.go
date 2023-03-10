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

import (
	"time"
)

// JobResult the result of the execution of a job
// note: ensure it is aligned with the same struct in piloth
type JobResult struct {
	// the unique job id
	JobId int64
	// indicates of the job was successful
	Success bool
	// the execution log for the job
	Log string
	// the error if any
	Err string
	// the completion time
	Time time.Time
}
