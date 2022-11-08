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
