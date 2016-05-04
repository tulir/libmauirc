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
	"strings"
)

// Handler is an IRC event handler
type Handler func(evt *irc.Message)

// AddHandler adds the given handler to the given code and returns the handler index
func (c *Connection) AddHandler(code string, handler Handler) int {
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

// RemoveHandler removes the handler with the given index from the given code
func (c *Connection) RemoveHandler(code string, index int) {
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

// GetHandlers gets all the handlers for the given code
func (c *Connection) GetHandlers(code string) (handlers []Handler, ok bool) {
	handlers, ok = c.handlers[code]
	return
}

// RunHandlers runs the handlers for the given code with the given event
func (c *Connection) RunHandlers(evt *irc.Message) {
	if tag, text, ok := ctcp.Decode(evt.Trailing); ok {
		evt.Command = fmt.Sprintf("CTCP_%s", tag)
		evt.Trailing = text
	}
	for _, handle := range c.handlers[evt.Command] {
		handle(evt)
	}
}
