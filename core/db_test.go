/*
  pilot control service
  © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package core

import (
	"testing"
)

func TestToHStoreString(t *testing.T) {
	m := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	s := toHStoreString(m)
	s1 := "\"key1\"=>\"value1\", \"key2\"=>\"value2\", \"key3\"=>\"value3\""
	if s != s1 {
		t.FailNow()
	}
}
