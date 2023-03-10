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
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"os"
	"path/filepath"
	os2 "southwinds.dev/os"
	"testing"
)

// converts metrics in open telemetry format to flat data points
func TestDbConv(t *testing.T) {
	pb := pmetric.ProtoUnmarshaler{}
	c := NewOtelDataPointConverter()
	entries, _ := os.ReadDir("test-data")
	for _, entry := range entries {
		content, _ := os.ReadFile(filepath.Join("test-data", entry.Name()))
		files, err := os2.ReadFileBatchFromBytes(content)
		if err != nil {
			t.Fatalf(err.Error())
		}
		for _, file := range files {
			metrics, err := pb.UnmarshalMetrics(file)
			if err != nil {
				t.Fatalf(err.Error())
			}
			points, err := c.Convert(metrics)
			if err != nil {
				t.Fatalf(err.Error())
			}
			for _, point := range points {
				b, _ := json.Marshal(point)
				fmt.Println(string(b[:]))
			}
		}
	}
}
