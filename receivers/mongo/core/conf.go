/*
  pilot control service - mongo event receiver
  © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package core

import "os"

// getDbConnString get the connection string to the MongoDb database
// e.g. mongodb://localhost:27017
// e.g. mongodb://user:password@127.0.0.1:27017/dbname?keepAlive=true&poolSize=30&autoReconnect=true&socketTimeoutMS=360000&connectTimeoutMS=360000
func getDbConnString() string {
	value := os.Getenv("PILOT_CTL_EVR_MONGO_CONN")
	if len(value) == 0 {
		panic("PILOT_CTL_EVR_MONGO_CONN not defined")
	}
	return value
}
