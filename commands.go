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
	"github.com/sorcix/irc"
	"github.com/sorcix/irc/ctcp"
	"strconv"
	"time"
)

// Send the given irc.Message
func (c *Connection) Send(msg *irc.Message) {
	c.output <- msg
}

// Action sends the given message to the given channel as a CTCP action message
func (c *Connection) Action(channel, msg string) {
	c.Privmsg(channel, ctcp.Action(msg))
}

// Privmsg sends the given message to the given channel
func (c *Connection) Privmsg(channel, msg string) {
	c.Send(&irc.Message{
		Command: irc.PRIVMSG,
		Params:  []string{channel, msg},
	})
}

// SendUser sends the USER message to the server
func (c *Connection) SendUser() {
	c.Send(&irc.Message{
		Command: irc.USER,
		Params:  []string{c.User, "0.0.0.0", "0.0.0.0", c.RealName},
	})
}

// SetNick updates the nick locally and sends a nick change request to the server
func (c *Connection) SetNick(nick string) {
	c.PreferredNick = nick
	c.Nick = nick
	c.Send(&irc.Message{
		Command: irc.NICK,
		Params:  []string{nick},
	})
}

// Ping the IRC server
func (c *Connection) Ping() {
	c.Send(&irc.Message{
		Command: irc.PING,
		Params:  []string{strconv.FormatInt(time.Now().UnixNano(), 10)},
	})
}
