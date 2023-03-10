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
	"log"
	"net/http"
	"southwinds.dev/pilotctl/types"
	"time"
)

type EventPublisher struct {
	conf *types.EventReceivers
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{conf: types.NewEventPubConf()}
}

func (p *EventPublisher) Publish(payload *types.Events) {
	if p.conf != nil {
		// loop through each configured events receiver
		for _, store := range p.conf.EventReceivers {
			// call retry publish function asynchronously
			go retryPublish(pubInput{
				payload: payload,
				uri:     store.URI,
				user:    store.User,
				pwd:     store.Pwd,
			})
		}
	} else {
		fmt.Printf("ERROR: no event publishers have been registered, events will be discarded\n")
	}
}

func retryPublish(input interface{}) {
	// retry 3 times applying exponential back-off intervals starting with 30 secs
	// adds jitter to the interval to prevent creating a Thundering Herd
	if err := Retry(3, 30*time.Second, publish, input); err != nil {
		// if the retry failed log the error
		log.Printf("ERROR: %s; events will be discarded\n", err)
	}
}

// publish the events to a specific receiver
// called asynchronously by the publisher
func publish(input interface{}) error {
	i := input.(pubInput)
	// create the connection configuration
	cfg := &ClientConf{
		BaseURI:            i.uri,
		Username:           i.user,
		Password:           i.pwd,
		InsecureSkipVerify: true,
		Timeout:            60 * time.Second,
	}
	client, err := NewClient(cfg)
	if err != nil {
		// client configuration error therefore it does not retry
		return StopRetry{fmt.Errorf("failed to create http client: %s\n", err)}
	}
	// create a request processor with the http credentials
	p := &processor{cfg: cfg}
	// post payload to receiver
	resp, err := client.Post(i.uri, i.payload, p.addToken)
	if err != nil {
		// server did not return an error but the operation failed therefore it keeps retrying
		return fmt.Errorf("failed to post events to receiver '%s': %s\n", i.uri, err)
	}
	if resp.StatusCode > 299 {
		// server returned error therefore it does not retry
		return StopRetry{fmt.Errorf("failed to post events to receiver '%s': '%s'\n", i.uri, resp.Status)}
	}
	// no error so it does not retry
	return nil
}

type pubInput struct {
	payload        *types.Events
	uri, user, pwd string
}

// MANAGE AUTHENTICATION:

// wrapper holding authentication information for the http processor function addToken
type processor struct {
	cfg *ClientConf
}

// add a basic authentication header to the http request
func (p *processor) addToken(req *http.Request, payload Serializable) error {
	// if authentication credentials are available
	if len(p.cfg.Username) > 0 && len(p.cfg.Password) > 0 {
		// add an authentication token to the request
		req.Header.Set("Authorization", basicAuthToken(p.cfg.Username, p.cfg.Password))
	}
	// all content type should be in JSON format
	req.Header.Set("Content-Type", "application/json")
	return nil
}
