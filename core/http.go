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
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func makeRequest(uri, method, user, pwd string, body io.Reader) ([]byte, error) {
	// create an http client
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		// set the client timeout period
		Timeout: 1 * time.Minute,
	}
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", basicAuthToken(user, pwd))
	// submits the request
	resp, err := client.Do(req)
	// check for error first
	if err != nil {
		return nil, err
	}
	// do we have a nil response?
	if resp == nil {
		return nil, errors.New(fmt.Sprintf("error: response was empty for resource: %s, check the service is up and running", uri))
	}
	// check for response status
	if resp.StatusCode >= 300 {
		return nil, errors.New(fmt.Sprintf("error: response returned status: %s", resp.Status))
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
