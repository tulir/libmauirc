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

// Tunnel contains functions to wrap IRC commands
type Tunnel interface {
	Send(msg *irc.Message)
	Action(channel, msg string)
	Privmsg(channel, msg string)
	Notice(channel, msg string)
	Away(msg string)
	RemoveAway()
	Invite(user, ch string)
	Kick(ch, user, msg string)
	Mode(target, flags, args string)
	Oper(username, password string)
	SetNick(nick string)
	Join(chs, keys string)
	Part(ch, msg string)
	List()
	Topic(ch, topic string)
	Whois(name string)
	Whowas(name string)
	Who(name string, op bool)
	Quit()
}

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
		Command:  irc.PRIVMSG,
		Params:   []string{channel},
		Trailing: msg,
	})
}

// Notice sends the given message to the given channel as a NOTICE
func (c *Connection) Notice(channel, msg string) {
	c.Send(&irc.Message{
		Command:  irc.NOTICE,
		Params:   []string{channel},
		Trailing: msg,
	})
}

// Away sets the away message
func (c *Connection) Away(msg string) {
	c.Send(&irc.Message{
		Command:  irc.AWAY,
		Trailing: msg,
	})
}

// RemoveAway removes the away status
func (c *Connection) RemoveAway() {
	c.Away("")
}

// Invite the given user to the given channel
func (c *Connection) Invite(user, ch string) {
	c.Send(&irc.Message{
		Command: irc.INVITE,
		Params:  []string{user, ch},
	})
}

// Kick the given user from the given channel with the given message
func (c *Connection) Kick(ch, user, msg string) {
	c.Send(&irc.Message{
		Command:  irc.KICK,
		Params:   []string{ch, user},
		Trailing: msg,
	})
}

// Mode changes channel and user modes
func (c *Connection) Mode(target, flags, args string) {
	c.Send(&irc.Message{
		Command: irc.MODE,
		Params:  []string{target, flags, args},
	})
}

// Oper authenticates the user as a server operator
func (c *Connection) Oper(username, password string) {
	c.Send(&irc.Message{
		Command: irc.OPER,
		Params:  []string{username, password},
	})
}

// SendUser sends the USER message to the server
func (c *Connection) SendUser() {
	c.Send(&irc.Message{
		Command:  irc.USER,
		Params:   []string{c.User, "0.0.0.0", "0.0.0.0"},
		Trailing: c.RealName,
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

// Pong replies to a Ping
func (c *Connection) Pong(msg string) {
	c.Send(&irc.Message{
		Command:  irc.PONG,
		Trailing: msg,
	})
}

// Join a channel
func (c *Connection) Join(chs string, keys string) {
	c.Send(&irc.Message{
		Command: irc.JOIN,
		Params:  []string{chs, keys},
	})
}

// Part a channel
func (c *Connection) Part(ch, msg string) {
	c.Send(&irc.Message{
		Command:  irc.JOIN,
		Params:   []string{ch},
		Trailing: msg,
	})
}

// List requests the server for a list of channels
func (c *Connection) List() {
	c.Send(&irc.Message{
		Command: irc.LIST,
	})
}

// Topic sets the topic of the given channel
func (c *Connection) Topic(ch, topic string) {
	c.Send(&irc.Message{
		Command:  irc.TOPIC,
		Params:   []string{ch},
		Trailing: topic,
	})
}

// Whois sends a WHOIS request on the given name
func (c *Connection) Whois(name string) {
	c.Send(&irc.Message{
		Command: irc.WHOIS,
		Params:  []string{name},
	})
}

// Whowas sends a WHOWAS request on the given name
func (c *Connection) Whowas(name string) {
	c.Send(&irc.Message{
		Command: irc.WHOWAS,
		Params:  []string{name},
	})
}

// Who sends a WHO request with the given name
func (c *Connection) Who(name string, op bool) {
	if op {
		c.Send(&irc.Message{
			Command: irc.WHOIS,
			Params:  []string{name, "o"},
		})
	} else {
		c.Send(&irc.Message{
			Command: irc.WHOIS,
			Params:  []string{name},
		})
	}
}

// Quit from the server
func (c *Connection) Quit() {
	c.Send(&irc.Message{
		Command:  irc.QUIT,
		Trailing: c.QuitMsg,
	})
	c.stopped = true
	c.quit = true
}
