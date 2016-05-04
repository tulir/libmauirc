// libmauirc - An IRC connection library for mauIRCd
// Copyright (C) 2016 Tulir Asokan

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package mauirc is the main package of this library
package mauirc

import (
	"fmt"
)

// SendRawf formats the given string and sends it to the IRC server
func (c *Connection) SendRawf(msg string, args ...interface{}) {
	c.SendRawString(fmt.Sprintf(msg, args...))
}

// SendRaw joins the given parts and sends the result to the IRC server
func (c *Connection) SendRaw(msg ...interface{}) {
	c.SendRawString(fmt.Sprint(msg...))
}

// SendRawString sends the given message to the IRC server
func (c *Connection) SendRawString(msg string) {

}
