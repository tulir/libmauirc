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

// Package libmauirc is the main package of this library
package libmauirc

import (
	"fmt"
	"github.com/sorcix/irc"
)

// AuthHandler handles authentication to an IRC server
type AuthHandler interface {
	Do(c *ConnImpl)
}

// PasswordAuth is an AuthHandler that authenticates using the IRC PASS command
type PasswordAuth struct {
	Password string
}

// Do the authentication
func (auth *PasswordAuth) Do(c *ConnImpl) {
	c.Send(&irc.Message{
		Command: irc.PASS,
		Params:  []string{auth.Password},
	})
}

// NickServAuth is an AuthHandler that authenticates with NickServ
type NickServAuth struct {
	Password string
}

// Do the authentication
func (auth *NickServAuth) Do(c *ConnImpl) {
	c.Privmsg("NickServ", fmt.Sprintf("IDENTIFY %s", auth.Password))
}
