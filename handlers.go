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
	"github.com/sorcix/irc/ctcp"
	"strconv"
	"strings"
	"time"
)

// HandlerHandler is a handler that handles handlers
type HandlerHandler interface {
	// AddHandler adds the given handler to the given code and returns the handler index
	AddHandler(code string, handler Handler) int
	// GetHandlers gets all the handlers for the given code
	GetHandlers(code string) (handlers []Handler, ok bool)
	// RunHandlers runs the handlers for the given code with the given event
	RunHandlers(evt *irc.Message)
}

// Handler is an IRC event handler
type Handler func(evt *irc.Message)

func (c *ConnImpl) AddHandler(code string, handler Handler) int {
	code = strings.ToUpper(code)
	handlers, ok := c.handlers[code]
	if !ok {
		handlers = make([]Handler, 1)
		handlers[0] = handler
		c.handlers[code] = handlers
		return 0
	}
	handlers = append(handlers, handler)
	c.handlers[code] = handlers
	return len(handlers) - 1
}

func (c *ConnImpl) RemoveHandler(code string, index int) {
	handlers, ok := c.handlers[code]
	if !ok || len(handlers) == 0 || len(handlers) >= index || index < 0 {
		return
	}
	if len(handlers) == 1 {
		delete(c.handlers, code)
	} else if index == 0 {
		handlers = handlers[1:]
	} else if index == len(handlers)-1 {
		handlers = handlers[:len(handlers)-1]
	} else {
		handlers = append(handlers[:index], handlers[index+1:]...)
	}
	c.handlers[code] = handlers
}

func (c *ConnImpl) GetHandlers(code string) (handlers []Handler, ok bool) {
	handlers, ok = c.handlers[code]
	return
}

func (c *ConnImpl) RunHandlers(evt *irc.Message) {
	if tag, text, ok := ctcp.Decode(evt.Trailing); ok && evt.Command == irc.PRIVMSG {
		evt.Command = fmt.Sprintf("CTCP_%s", tag)
		evt.Trailing = text
	}
	evt.Params = append(evt.Params, strings.Split(evt.Trailing, " ")...)
	for _, handle := range c.handlers[evt.Command] {
		handle(evt)
	}
}

// AddStdHandlers add standard IRC handlers for this connection
// The standard handlers include an IRC ERROR handler, ping and pong handler, CTCP version, userinfo, clientinfo,
// time and ping handlers and a nick change handler.
func (c *ConnImpl) AddStdHandlers() {
	c.AddHandler("ERROR", func(evt *irc.Message) {
		c.Disconnect()
	})

	c.AddHandler("PING", func(evt *irc.Message) {
		c.Pong(evt.Trailing)
	})

	c.AddHandler("PONG", func(evt *irc.Message) {
		ns, _ := strconv.ParseInt(evt.Trailing, 10, 64)
		delta := time.Duration(time.Now().UnixNano() - ns)
		c.Debugfln("Lag: %v", delta)
	})

	c.AddHandler("CTCP_VERSION", func(evt *irc.Message) {
		c.Send(&irc.Message{
			Command:  "NOTICE",
			Params:   []string{evt.Name},
			Trailing: ctcp.Version(c.Version),
		})
	})

	c.AddHandler("CTCP_USERINFO", func(evt *irc.Message) {
		c.Send(&irc.Message{
			Command:  "NOTICE",
			Params:   []string{evt.Name},
			Trailing: ctcp.UserInfo(c.User),
		})
	})

	c.AddHandler("CTCP_CLIENTINFO", func(evt *irc.Message) {
		c.Send(&irc.Message{
			Command:  "NOTICE",
			Params:   []string{evt.Name},
			Trailing: ctcp.ClientInfo("CLIENTINFO PING VERSION TIME USERINFO CLIENTINFO"),
		})
	})

	c.AddHandler("CTCP_TIME", func(evt *irc.Message) {
		c.Send(&irc.Message{
			Command:  "NOTICE",
			Params:   []string{evt.Name},
			Trailing: ctcp.TimeReply(),
		})
	})

	c.AddHandler("CTCP_PING", func(evt *irc.Message) {
		c.Send(&irc.Message{
			Command:  "NOTICE",
			Params:   []string{evt.Name},
			Trailing: ctcp.Ping(evt.Trailing),
		})
	})

	nickused := func(evt *irc.Message) {
		if len(c.Nick) >= 9 {
			c.SetNick("_" + c.Nick)
		} else {
			c.SetNick(c.Nick + "_")
		}
	}
	c.AddHandler("437", nickused)
	c.AddHandler("433", nickused)

	c.AddHandler("NICK", func(evt *irc.Message) {
		if evt.Name == c.PreferredNick {
			c.Nick = evt.Trailing
		}
	})

	c.AddHandler("001", func(evt *irc.Message) {
		c.Nick = evt.Params[0]
	})
}
