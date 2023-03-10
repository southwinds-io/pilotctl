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

package types

import (
	"bytes"
	"fmt"
	"southwinds.dev/artisan/data"
	"southwinds.dev/artisan/merge"
)

// CmdInfo all the information required by pilot to execute a command
type CmdInfo struct {
	JobId         int64       `json:"job_id"`
	Package       string      `json:"package"`
	Function      string      `json:"function"`
	User          string      `json:"user"`
	Pwd           string      `json:"pwd"`
	Verbose       bool        `json:"verbose"`
	Containerised bool        `json:"containerised"`
	Input         *data.Input `json:"input,omitempty"`
}

func (c *CmdInfo) Value() string {
	var artCmd string
	// if command is to run in a runtime
	if c.Containerised {
		// use art exec
		artCmd = "exec"
	} else {
		// otherwise, use art exe
		artCmd = "exe"
	}
	// if user credentials for the Artisan registry were provided
	if len(c.User) > 0 && len(c.Pwd) > 0 {
		// pass the credentials to the art cli
		return fmt.Sprintf("art %s -u %s:%s %s %s", artCmd, c.User, c.Pwd, c.Package, c.Function)
	}
	// otherwise run the command without credentials (assume public registry)
	return fmt.Sprintf("art %s %s %s", artCmd, c.Package, c.Function)
}

func (c *CmdInfo) Env() []string {
	var vars []string
	// append vars
	for _, v := range c.Input.Var {
		vars = append(vars, fmt.Sprintf("%s=%s", v.Name, v.Value))
	}
	// append secrets
	for _, s := range c.Input.Secret {
		vars = append(vars, fmt.Sprintf("%s=%s", s.Name, s.Value))
	}
	return vars
}

func (c *CmdInfo) Envar() *merge.Envar {
	return merge.NewEnVarFromSlice(c.Env())
}

func (c *CmdInfo) PrintEnv() string {
	var vars bytes.Buffer
	vars.WriteString("printing variables passed to the shell\n{\n")
	for _, v := range c.Input.Var {
		vars.WriteString(fmt.Sprintf("%s=%s\n", v.Name, v.Value))
	}
	vars.WriteString("}\n")
	return vars.String()
}
