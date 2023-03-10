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
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSubmitMetrics(t *testing.T) {
	os.Setenv("PILOT_CTL_DB_HOST", "localhost")
	os.Setenv("PILOT_CTL_DB_USER", "pilotctl")
	os.Setenv("PILOT_CTL_DB_PWD", "p1l0tctl")
	os.Setenv("PILOT_CTL_ILINK_URI", "http://localhost:8080")
	os.Setenv("PILOT_CTL_ILINK_USER", "admin")
	os.Setenv("PILOT_CTL_ILINK_PWD", "0n1x")
	os.Setenv("PILOT_CTL_ILINK_INSECURE_SKIP_VERIFY", "true")
	os.Setenv("PILOT_CTL_TELEM_CONN", "test:mongo-metrics")

	// pravega connector  settings
	// os.Setenv("PRAVEGA_CONN_HOST", "localhost:9090")
	// os.Setenv("PRAVEGA_CONN_SCOPE", "host")
	// os.Setenv("PRAVEGA_CONN_STREAM", "metrics")

	// mongo connector settings
	os.Setenv("MONGO_CONN_CONNECTION_STRING", "mongodb://root:example@localhost:27017")
	os.Setenv("MONGO_CONN_DATABASE", "metrics")
	os.Setenv("MONGO_CONN_METRICS_COLLECTION", "host")

	api, err := NewAPI(NewConf())
	if err != nil {
		t.Fatalf(err.Error())
	}
	entries, _ := os.ReadDir("test-data")
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		abs, _ := filepath.Abs(filepath.Join("test-data", entry.Name()))
		content, err := os.ReadFile(abs)
		if err != nil {
			t.Fatalf(err.Error())
		}
		r := api.SubmitMetrics("test", content)
		fmt.Println(r.Error)
	}
}
