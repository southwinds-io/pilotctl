/*
pilot control service
© 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
Contributors to this project, hereby assign copyright in this code to the project,
to be licensed under the same terms as the rest of the code.
*/
package core

import (
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"os"
	"path/filepath"
	"testing"
)

// converts metrics in open telemetry format to flat data points
func TestDbConv(t *testing.T) {
	pb := pmetric.ProtoUnmarshaler{}
	c := NewOtelDataPointConverter()
	entries, _ := os.ReadDir("test-data")
	for _, entry := range entries {
		content, _ := os.ReadFile(filepath.Join("test-data", entry.Name()))
		metrics, _ := pb.UnmarshalMetrics(content)
		points, _ := c.Convert(metrics)
		for _, point := range points {
			b, _ := json.Marshal(point)
			fmt.Println(string(b[:]))
		}
	}
}
