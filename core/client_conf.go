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
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ClientConf client configuration information
type ClientConf struct {
	// the base URI for the service
	BaseURI string
	// disables TLS certificate verification
	InsecureSkipVerify bool
	// the user username for Basic and OpenId user authentication
	Username string
	// the user password for Basic and OpenId user authentication
	Password string
	// time out
	Timeout time.Duration
	// client proxy
	Proxy func(*http.Request) (*url.URL, error)
}

// gets the authentication token based on the authentication mode selected
func (cfg *ClientConf) getAuthToken() (string, error) {
	return cfg.basicToken(cfg.Username, cfg.Password), nil
}

// validates the client configuration
func checkConf(cfg *ClientConf) error {
	if len(cfg.BaseURI) == 0 {
		return errors.New("BaseURI is not defined")
	}
	// if the protocol is not specified, the add http as default
	// this is to avoid the server producing empty responses if no protocol is specified in the URI
	if !strings.HasPrefix(strings.ToLower(cfg.BaseURI), "http") {
		log.Warn().Msgf("no protocol defined for Onix URI '%s', 'http://' will be added to it", cfg.BaseURI)
		cfg.BaseURI = fmt.Sprintf("http://%s", cfg.BaseURI)
	}

	if len(cfg.Username) == 0 {
		return errors.New("username is not defined")
	}
	if len(cfg.Password) == 0 {
		return errors.New("password is not defined")
	}

	// if timeout is zero, it never timeout so is not good
	if cfg.Timeout == 0*time.Second {
		// set a default timeout of 30 secs
		cfg.Timeout = 30 * time.Second
	}
	return nil
}

// creates a new Basic Authentication Token
func (cfg *ClientConf) basicToken(user string, pwd string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pwd))))
}
