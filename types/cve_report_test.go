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
	"fmt"
	"os"
	"testing"
)

func TestLoad2(t *testing.T) {
	data, err := os.ReadFile("cve/sample_report.json")
	if err != nil {
		t.Fatal(fmt.Errorf("failed to read report: %s", err.Error()))
	}
	r, err := NewCveReport(data)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("medium %d\n", r.Medium())
	fmt.Printf("low %d\n", r.Low())
	for _, cve := range r.Cves {
		fmt.Printf("CVE: %s (fixed: %t)\n", cve.Id, cve.Fixed())
	}
}
