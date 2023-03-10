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

package core

import (
	"math"
	"time"
)

func toElapsedValues(rfc850time string) (int, string, error) {
	if len(rfc850time) == 0 {
		return 0, "", nil
	}
	created, err := time.Parse(time.RFC850, rfc850time)
	if err != nil {
		return 0, "", err
	}
	elapsed := time.Since(created)
	seconds := elapsed.Seconds()
	minutes := elapsed.Minutes()
	hours := elapsed.Hours()
	days := hours / 24
	weeks := days / 7
	months := weeks / 4
	years := months / 12

	if math.Trunc(years) > 0 {
		return int(years), "y", nil
	} else if math.Trunc(months) > 0 {
		return int(months), "M", nil
	} else if math.Trunc(weeks) > 0 {
		return int(weeks), "w", nil
	} else if math.Trunc(days) > 0 {
		return int(days), "d", nil
	} else if math.Trunc(hours) > 0 {
		return int(hours), "H", nil
	} else if math.Trunc(minutes) > 0 {
		return int(minutes), "m", nil
	}
	return int(seconds), "s", nil
}
