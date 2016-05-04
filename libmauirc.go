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
	"crypto/tls"
	"fmt"
	"github.com/sorcix/irc"
	"io"
	"net"
	"sync"
	"time"
)

// Version is the IRC client version string
var Version = "libmauirc 0.1"

// Connection is an IRC connection
type Connection struct {
	sync.WaitGroup
	PingFreq  time.Duration
	KeepAlive time.Duration
	Timeout   time.Duration
	prevMsg   time.Time

	Version       string
	PreferredNick string
	Nick          string
	User          string
	RealName      string
	QuitMsg       string

	handlers map[string][]Handler

	Auths   []AuthHandler
	Address AddressHandler

	DebugWriter io.Writer
	stopped     bool
	quit        bool
	UseTLS      bool
	TLSConfig   *tls.Config
	socket      net.Conn
	output      chan *irc.Message
	Errors      chan error
	end         chan struct{}
}

// Create an IRC connection
func Create(nick, user, realname string, addr AddressHandler) *Connection {
	if len(realname) == 0 {
		realname = user
	}
	return &Connection{
		Nick:          nick,
		PreferredNick: nick,
		User:          user,
		RealName:      realname,
		Address:       addr,
		Auths:         make([]AuthHandler, 0),
		end:           make(chan struct{}),
		Version:       Version,
		KeepAlive:     4 * time.Minute,
		Timeout:       1 * time.Minute,
		PingFreq:      15 * time.Minute,
		QuitMsg:       Version,
	}
}

// Connect to the IRC server
func (c *Connection) Connect() error {
	c.stopped = true

	if c.Address == nil {
		return ErrInvalidAddress
	} else if len(c.Nick) == 0 {
		return ErrInvalidNick
	} else if len(c.User) == 0 {
		return ErrInvalidUser
	} else if len(c.RealName) == 0 {
		return ErrInvalidRealName
	}

	var err error
	if c.UseTLS {
		dialer := &net.Dialer{Timeout: c.Timeout}
		c.socket, err = tls.DialWithDialer(dialer, "tcp", c.Address.String(), c.TLSConfig)
	} else {
		c.socket, err = net.DialTimeout("tcp", c.Address.String(), c.Timeout)
	}
	if err != nil {
		c.Debugfln("Failed to connect to %s: %v", c.Address.String(), err)
		return ConnectionError{Cause: err}
	}
	c.Debugfln("Successfully connected to %s (%s)", c.Address.String(), c.socket.RemoteAddr().String())

	c.stopped = false

	c.output = make(chan *irc.Message, 10)
	c.Errors = make(chan error, 2)
	c.Add(3)

	go c.readLoop()
	go c.writeLoop()
	go c.pingLoop()

	for _, auth := range c.Auths {
		auth.Do(c)
	}

	c.SetNick(c.Nick)
	c.SendUser()
	return nil
}

// Debugf prints a debug message with fmt.Fprintf
func (c *Connection) Debugf(msg string, args ...interface{}) {
	if c.DebugWriter != nil {
		fmt.Fprintf(c.DebugWriter, msg, args...)
	}
}

// Debugfln prints a debug message with fmt.Fprintf and appends \n
func (c *Connection) Debugfln(msg string, args ...interface{}) {
	if c.DebugWriter != nil {
		fmt.Fprintf(c.DebugWriter, msg, args...)
		fmt.Fprint(c.DebugWriter, "\n")
	}
}

// Debug prints a debug message with fmt.Fprint
func (c *Connection) Debug(parts ...interface{}) {
	if c.DebugWriter != nil {
		fmt.Fprint(c.DebugWriter, parts...)
	}
}

// Debugln prints a debug message with fmt.Fprintln
func (c *Connection) Debugln(parts ...interface{}) {
	if c.DebugWriter != nil {
		fmt.Fprintln(c.DebugWriter, parts...)
	}
}
