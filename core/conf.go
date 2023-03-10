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
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type ConfKey string

const (
	ConfDbName                  ConfKey = "PILOT_CTL_DB_NAME"
	ConfDbHost                  ConfKey = "PILOT_CTL_DB_HOST"
	ConfDbPort                  ConfKey = "PILOT_CTL_DB_PORT"
	ConfDbUser                  ConfKey = "PILOT_CTL_DB_USER"
	ConfDbPwd                   ConfKey = "PILOT_CTL_DB_PWD"
	ConfPingIntervalSecs        ConfKey = "PILOT_CTL_PING_INTERVAL_SECS"
	ConfILinkUri                ConfKey = "PILOT_CTL_ILINK_URI"
	ConfILinkUser               ConfKey = "PILOT_CTL_ILINK_USER"
	ConfILinkPwd                ConfKey = "PILOT_CTL_ILINK_PWD"
	ConfILinkInsecureSkipVerify ConfKey = "PILOT_CTL_ILINK_INSECURE_SKIP_VERIFY"
	ConfArtRegURI               ConfKey = "PILOT_CTL_ART_REG_URI"
	ConfArtRegUser              ConfKey = "PILOT_CTL_ART_REG_USER"
	ConfArtRegPwd               ConfKey = "PILOT_CTL_ART_REG_PWD"
	ConfArtRegPackageFilter     ConfKey = "PILOT_CTL_ART_REG_PACKAGE_FILTER"
	ConfActURI                  ConfKey = "PILOT_CTL_ACTIVATION_URI"
	ConfActUser                 ConfKey = "PILOT_CTL_ACTIVATION_USER"
	ConfActPwd                  ConfKey = "PILOT_CTL_ACTIVATION_PWD"
	ConfTenant                  ConfKey = "PILOT_CTL_TENANT"
	ConfDbMaxConn               ConfKey = "PILOT_CTL_DB_MAXCONN"
	ConfCorsOrigin              ConfKey = "PILOT_CTL_CORS_ORIGIN"
	ConfCorsHeaders             ConfKey = "PILOT_CTL_CORS_HEADERS"
	ConfSyncPath                ConfKey = "PILOT_CTL_SYNC_PATH"
	ConfTelemBufferPath         ConfKey = "PILOT_CTL_TELEM_BUFFER_PATH"
	ConfTelemConnectors         ConfKey = "PILOT_CTL_TELEM_CONN"
)

type Conf struct {
}

func NewConf() *Conf {
	return &Conf{}
}

func (c *Conf) get(key ConfKey) string {
	return os.Getenv(string(key))
}

func (c *Conf) getSyncPath() string {
	return os.Getenv(string(ConfSyncPath))
}

func (c *Conf) getTelemBufferPath() string {
	return os.Getenv(string(ConfTelemBufferPath))
}

func (c *Conf) getDbName() string {
	value := os.Getenv(string(ConfDbName))
	if len(value) == 0 {
		return "pilotctl"
	}
	return value
}

func (c *Conf) getDbHost() string {
	return c.getValue(ConfDbHost)
}

func (c *Conf) getDbPort() string {
	value := os.Getenv(string(ConfDbPort))
	if len(value) == 0 {
		return "5432"
	}
	return value
}

func (c *Conf) getDbUser() string {
	return c.getValue(ConfDbUser)
}

func (c *Conf) getDbPwd() string {
	return c.getValue(ConfDbPwd)
}

func (c *Conf) GetTenant() string {
	return c.getValue(ConfTenant)
}

func (c *Conf) GetActivationURI() string {
	return c.getValue(ConfActURI)
}

func (c *Conf) GetActivationUser() string {
	return c.getValue(ConfActUser)
}

func (c *Conf) GetActivationPwd() string {
	return c.getValue(ConfActPwd)
}

// PingIntervalSecs the pilot ping interval
func (c *Conf) PingIntervalSecs() time.Duration {
	defaultValue, _ := time.ParseDuration("15s")
	value := os.Getenv(string(ConfPingIntervalSecs))
	if len(value) == 0 {
		return defaultValue
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		fmt.Printf("WARNING: %s is invalid, defaulting to %d\n", ConfPingIntervalSecs, defaultValue)
		return defaultValue
	}
	interval, err := time.ParseDuration(fmt.Sprintf("%ds", v))
	if err != nil {
		fmt.Printf("WARNING: %s is invalid, defaulting to %d\n", ConfPingIntervalSecs, defaultValue)
		return defaultValue
	}
	return interval
}

func (c *Conf) getOxWapiUrl() string {
	return c.getValue(ConfILinkUri)
}

func (c *Conf) getOxWapiUsername() string {
	return c.getValue(ConfILinkUser)
}

func (c *Conf) getOxWapiPassword() string {
	return c.getValue(ConfILinkPwd)
}

func (c *Conf) getArtRegUri() string {
	return c.getValue(ConfArtRegURI)
}

func (c *Conf) getArtRegUser() string {
	return c.getValue(ConfArtRegUser)
}

func (c *Conf) getArtRegPwd() string {
	return c.getValue(ConfArtRegPwd)
}

func (c *Conf) getValue(key ConfKey) string {
	value := os.Getenv(string(key))
	if len(value) == 0 {
		fmt.Printf("ERROR: variable %s not defined", key)
		os.Exit(1)
	}
	return value
}

func (c *Conf) getValueWithError(key ConfKey) (string, error) {
	value := os.Getenv(string(key))
	if len(value) == 0 {
		err := fmt.Sprintf("WARNING: variable %s not defined", key)
		log.Printf(err)
		return "", errors.New(err)
	}
	return value, nil
}

func (c *Conf) getOxWapiInsecureSkipVerify() bool {
	b, err := strconv.ParseBool(c.getValue(ConfILinkInsecureSkipVerify))
	if err != nil {
		fmt.Printf("ERROR: invalid value for variable %s", ConfILinkInsecureSkipVerify)
		os.Exit(1)
	}
	return b
}

func (c *Conf) getDbMaxConn() int {
	defaultMaxConn := 10
	value := os.Getenv(string(ConfDbMaxConn))
	if len(value) == 0 {
		return defaultMaxConn
	}
	maxConn, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("WARNING: failed to parse db max connections: %s, defaulting to %d\n", err, defaultMaxConn)
		return defaultMaxConn
	}
	return maxConn
}

func (c *Conf) GetCorsOrigin() string {
	value, err := c.getValueWithError(ConfCorsOrigin)
	if err != nil {
		log.Printf("WARNING: will not allow CORS")
	}
	return value
}

func (c *Conf) GetCorsHeaders() string {
	value, err := c.getValueWithError(ConfCorsHeaders)
	if err != nil {
		log.Printf("WARNING: will block headers such as Authorization and Origin on CORS OPTIONS requests")
	}
	return value
}

func (c *Conf) GetArtRegPackageFilter() string {
	return c.getValue(ConfArtRegPackageFilter)
}
