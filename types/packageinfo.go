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

// PackageInfo describes a package and all its tags
type PackageInfo struct {
	Name string    `json:"name"`
	Tags []TagInfo `json:"tags,omitempty"`
}

type TagInfo struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Ref     string `json:"ref"`
	Created string `json:"created"`
	Type    string `json:"type"`
	Size    string `json:"size"`
}
