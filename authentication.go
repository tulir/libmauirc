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
	"github.com/sorcix/irc"
)

// AuthHandler handles authentication to an IRC server
type AuthHandler interface {
	Do(c *Connection)
}

// PasswordAuth is an AuthHandler that authenticates using the IRC PASS command
type PasswordAuth struct {
	Password string
}

func (auth *PasswordAuth) Do(c *Connection) {
	c.Send(&irc.Message{
		Command: irc.PASS,
		Params:  []string{auth.Password},
	})
}

// NickServAuth is an AuthHandler that authenticates with NickServ
type NickServAuth struct {
	Password string
}

func (auth *NickServAuth) Do(c *Connection) {
	c.Privmsg("NickServ", fmt.Sprintf("IDENTIFY %s", auth.Password))
}
