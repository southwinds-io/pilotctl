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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"southwinds.dev/artisan/data"
)

var (
	api *API
)

func Api() *API {
	var (
		err    error
		newAPI *API
	)
	if api == nil {
		newAPI, err = NewAPI(new(Conf))
		if err != nil {
			log.Fatalf("ERROR: fail to create backend services API: %s", err)
		}
		api = newAPI
	}
	return api
}

func basicAuthToken(user, pwd string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pwd))))
}

// getInputFromMap transform an input in map format to input struct format
func getInputFromMap(inputMap map[string]interface{}) (*data.Input, error) {
	input := &data.Input{}
	in := inputMap["input"]
	if in != nil {
		// load vars
		vars := in.(map[string]interface{})["var"]
		if vars != nil {
			input.Var = data.Vars{}
			v := vars.([]interface{})
			for _, i := range v {
				varMap := i.(map[string]interface{})
				nameValue, ok := varMap["name"].(string)
				if !ok {
					return nil, fmt.Errorf("variable NAME must be provided, can't process payload\n")
				}
				descValue, ok := varMap["description"].(string)
				if !ok {
					descValue = ""
				}
				typeValue, ok := varMap["type"].(string)
				if !ok || len(typeValue) == 0 {
					typeValue = "string"
				}
				requiredValue, ok := varMap["required"].(bool)
				if !ok {
					requiredValue = false
				}
				valueValue, ok := varMap["value"].(string)
				if !ok && requiredValue {
					return nil, fmt.Errorf("variable %s VALUE not provided, can't process payload\n", nameValue)
				}
				vv := &data.Var{
					Name:        nameValue,
					Description: descValue,
					Value:       valueValue,
					Type:        typeValue,
					Required:    requiredValue,
				}
				input.Var = append(input.Var, vv)
			}
		}
		// load secrets
		secrets := in.(map[string]interface{})["secret"]
		if secrets != nil {
			input.Secret = data.Secrets{}
			v := secrets.([]interface{})
			for _, i := range v {
				varMap := i.(map[string]interface{})
				nameValue, ok := varMap["name"].(string)
				if !ok {
					return nil, fmt.Errorf("secret NAME must be provided, can't process payload\n")
				}
				descValue, ok := varMap["description"].(string)
				if !ok {
					descValue = ""
				}
				requiredValue, ok := varMap["required"].(bool)
				if !ok {
					requiredValue = false
				}
				// a secret might not contain value if not required
				var valueValue string
				if varMap["value"] != nil {
					valueValue = varMap["value"].(string)
				}
				vv := &data.Secret{
					Name:        nameValue,
					Description: descValue,
					Value:       valueValue,
					Required:    requiredValue,
				}
				input.Secret = append(input.Secret, vv)
			}
		}
	}
	return input, nil
}

func HttpRequest(client *http.Client, uri, method, user, pwd string, resultCode int) ([]byte, error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create http request to backend: %s\n", err)
	}
	req.Header.Add("Authorization", basicToken(user, pwd))
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request to backend failed: %s\n", err)
	}
	if resp.StatusCode != resultCode {
		return nil, fmt.Errorf("http request to backend failed: %s\n", resp.Status)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read backend service response: %s\n", err)
	}
	return respBody, nil
}

func basicToken(user string, pwd string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pwd))))
}
