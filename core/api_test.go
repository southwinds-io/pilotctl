package core

import (
	"fmt"
	"os"
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
	os.Setenv("PILOT_CTL_TELEM_CONN", "test:pravega-conn")
	os.Setenv("PRAVEGA_CONN_HOST", "localhost:9090")
	os.Setenv("PRAVEGA_CONN_SCOPE", "host")
	os.Setenv("PRAVEGA_CONN_STREAM", "metrics")
	api, err := NewAPI(NewConf())
	if err != nil {
		t.Fatalf(err.Error())
	}
	c, err := os.ReadFile("protobuf.pb")
	if err != nil {
		t.Fatalf(err.Error())
	}
	r := api.SubmitMetrics("test", c)
	fmt.Println(r.Error)
}
