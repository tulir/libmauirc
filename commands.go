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
	"strconv"
	"time"

	"github.com/sorcix/irc"
	"github.com/sorcix/irc/ctcp"
)

// Tunnel contains functions to wrap IRC commands
type Tunnel interface {
	// Send the given irc.Message
	Send(msg *irc.Message)
	// Action sends the given message to the given channel as a CTCP action message
	Action(channel, msg string)
	// Privmsg sends the given message to the given channel
	Privmsg(channel, msg string)
	// Notice sends the given message to the given channel as a NOTICE
	Notice(channel, msg string)
	// Away sets the away message
	Away(msg string)
	// RemoveAway removes the away status
	RemoveAway()
	// Invite the given user to the given channel
	Invite(user, ch string)
	// Kick the given user from the given channel with the given message
	Kick(ch, user, msg string)
	// Mode changes channel and user modes
	Mode(target, flags, args string)
	// Oper authenticates the user as a server operator
	Oper(username, password string)
	// SetNick updates the nick locally and sends a nick change request to the server
	SetNick(nick string)
	// Join a channel
	Join(chs, keys string)
	// Part a channel
	Part(ch, msg string)
	// List requests the server for a list of channels
	List()
	// Topic sets the topic of the given channel
	Topic(ch, topic string)
	// Whois sends a WHOIS request on the given name
	Whois(name string)
	// Whowas sends a WHOWAS request on the given name
	Whowas(name string)
	// Who sends a WHO request with the given name
	Who(name string, op bool)
	// Quit from the server
	Quit()
}

// Send - See Tunnel interface docs
func (c *ConnImpl) Send(msg *irc.Message) {
	c.output <- msg
}

// Action - See Tunnel interface docs
// Action - See Tunnel interface docs
func (c *ConnImpl) Action(channel, msg string) {
	c.Privmsg(channel, ctcp.Action(msg))
}

// Privmsg - See Tunnel interface docs
// Privmsg - See Tunnel interface docs
func (c *ConnImpl) Privmsg(channel, msg string) {
	c.Send(&irc.Message{
		Command:  irc.PRIVMSG,
		Params:   []string{channel},
		Trailing: msg,
	})
}

// Notice - See Tunnel interface docs
func (c *ConnImpl) Notice(channel, msg string) {
	c.Send(&irc.Message{
		Command:  irc.NOTICE,
		Params:   []string{channel},
		Trailing: msg,
	})
}

// Away - See Tunnel interface docs
func (c *ConnImpl) Away(msg string) {
	c.Send(&irc.Message{
		Command:  irc.AWAY,
		Trailing: msg,
	})
}

// RemoveAway - See Tunnel interface docs
func (c *ConnImpl) RemoveAway() {
	c.Away("")
}

// Invite - See Tunnel interface docs
func (c *ConnImpl) Invite(user, ch string) {
	c.Send(&irc.Message{
		Command: irc.INVITE,
		Params:  []string{user, ch},
	})
}

// Kick - See Tunnel interface docs
func (c *ConnImpl) Kick(ch, user, msg string) {
	c.Send(&irc.Message{
		Command:  irc.KICK,
		Params:   []string{ch, user},
		Trailing: msg,
	})
}

// Mode - See Tunnel interface docs
func (c *ConnImpl) Mode(target, flags, args string) {
	c.Send(&irc.Message{
		Command: irc.MODE,
		Params:  []string{target, flags, args},
	})
}

// Oper - See Tunnel interface docs
func (c *ConnImpl) Oper(username, password string) {
	c.Send(&irc.Message{
		Command: irc.OPER,
		Params:  []string{username, password},
	})
}

// SetNick - See Tunnel interface docs
func (c *ConnImpl) SetNick(nick string) {
	c.PreferredNick = nick
	c.Nick = nick
	c.Send(&irc.Message{
		Command: irc.NICK,
		Params:  []string{nick},
	})
}

// Join - See Tunnel interface docs
func (c *ConnImpl) Join(chs string, keys string) {
	c.Send(&irc.Message{
		Command: irc.JOIN,
		Params:  []string{chs, keys},
	})
}

// Part - See Tunnel interface docs
func (c *ConnImpl) Part(ch, msg string) {
	c.Send(&irc.Message{
		Command:  irc.PART,
		Params:   []string{ch},
		Trailing: msg,
	})
}

// List - See Tunnel interface docs
func (c *ConnImpl) List() {
	c.Send(&irc.Message{
		Command: irc.LIST,
	})
}

// Topic - See Tunnel interface docs
func (c *ConnImpl) Topic(ch, topic string) {
	c.Send(&irc.Message{
		Command:  irc.TOPIC,
		Params:   []string{ch},
		Trailing: topic,
	})
}

// Whois - See Tunnel interface docs
func (c *ConnImpl) Whois(name string) {
	c.Send(&irc.Message{
		Command: irc.WHOIS,
		Params:  []string{name},
	})
}

// Whowas - See Tunnel interface docs
func (c *ConnImpl) Whowas(name string) {
	c.Send(&irc.Message{
		Command: irc.WHOWAS,
		Params:  []string{name},
	})
}

// Who - See Tunnel interface docs
func (c *ConnImpl) Who(name string, op bool) {
	if op {
		c.Send(&irc.Message{
			Command: irc.WHO,
			Params:  []string{name, "o"},
		})
	} else {
		c.Send(&irc.Message{
			Command: irc.WHO,
			Params:  []string{name},
		})
	}
}

// Quit - See Tunnel interface docs
func (c *ConnImpl) Quit() {
	c.Send(&irc.Message{
		Command:  irc.QUIT,
		Trailing: c.QuitMsg,
	})
	c.Lock()
	c.quit = true
	c.Unlock()
}

// SendUser sends the USER message to the server
// SendUser - See Tunnel interface docs
func (c *ConnImpl) SendUser() {
	c.Send(&irc.Message{
		Command:  irc.USER,
		Params:   []string{c.User, "0.0.0.0", "0.0.0.0"},
		Trailing: c.RealName,
	})
}

// Ping the IRC server
// Ping - See Tunnel interface docs
func (c *ConnImpl) Ping() {
	c.Send(&irc.Message{
		Command: irc.PING,
		Params:  []string{strconv.FormatInt(time.Now().UnixNano(), 10)},
	})
}

// Pong replies to a Ping
func (c *ConnImpl) Pong(msg string) {
	c.Send(&irc.Message{
		Command:  irc.PONG,
		Trailing: msg,
	})
}
