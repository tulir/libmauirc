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
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/sorcix/irc"
)

// Version is the IRC client version string
var Version = "libmauirc 0.2"

// Debugger is something to send debug messages to
type Debugger interface {
	// Debug prints a debug message with fmt.Fprint
	Debug(parts ...interface{})
	// Debugln prints a debug message with fmt.Fprintln
	Debugln(parts ...interface{})
	// Debugf prints a debug message with fmt.Fprintf
	Debugf(msg string, args ...interface{})
	// Debugfln prints a debug message with fmt.Fprintf and appends \n
	Debugfln(msg string, args ...interface{})
	// SetDebugWriter changes the io.Writer to which debug data should be written to
	SetDebugWriter(writer io.Writer)
}

// Data has miscancellous functions to change IRC info.
type Data interface {
	GetNick() string
	GetPreferredNick() string
	SetQuitMessage(msg string)
	SetRealName(realname string)
	SetVersion(version string)
	SetUseTLS(tls bool)
	AddAuth(auth AuthHandler)
	SetAddress(addr Address)
}

// Connectable contains functions to connect and disconnect
type Connectable interface {
	// Connect to the server.
	// An error will be returned if some settings are incorrect or if an error is received while connecting.
	Connect() error
	// Loop to automatically reconnect to the server.
	Loop()
	// Disconnect from the server.
	Disconnect()
	// Connected checks if the connection is active.
	Connected() bool
	// LocalAddr gets the local address of a connection
	LocalAddr() net.Addr
}

// ErrorStream contains a function that returns a channel of non-lethal errors.
type ErrorStream interface {
	Errors() chan error
}

// Connection contains all the necessary interfaces for an IRC connection.
// The default implementation is ConnImpl.
type Connection interface {
	Debugger
	HandlerHandler
	Tunnel
	Data
	Connectable
	ErrorStream
}

// ConnImpl is the default implementation of Connection.
// The functions here don't have separate documentation. See the documentation of the interfaces
// Connection contains for documentation on ConnImpl's functions.
type ConnImpl struct {
	sync.Mutex
	sync.WaitGroup
	PingFreq             time.Duration
	KeepAlive            time.Duration
	Timeout              time.Duration
	AutoreconnectTimeout time.Duration
	prevMsg              time.Time

	Version       string
	PreferredNick string
	Nick          string
	User          string
	RealName      string
	QuitMsg       string
	Lag           int64

	handlers map[string][]Handler
	Auth     []AuthHandler
	Address  Address

	DebugWriter      io.Writer
	reconnectAttempt int
	stopped          bool
	quit             bool
	UseTLS           bool
	Autoreconnect    bool
	TLSConfig        *tls.Config
	socket           net.Conn
	output           chan *irc.Message
	errors           chan error
	disconnected     chan error
	end              chan interface{}
}

// Create an IRC connection with the given details.
// By default, RealName is set to the same value as user.
func Create(nick, user string, addr Address) Connection {
	c := &ConnImpl{
		Nick:                 nick,
		PreferredNick:        nick,
		User:                 user,
		RealName:             user,
		Address:              addr,
		Auth:                 make([]AuthHandler, 0),
		handlers:             make(map[string][]Handler),
		Version:              Version,
		KeepAlive:            4 * time.Minute,
		AutoreconnectTimeout: 7 * time.Minute,
		Autoreconnect:        true,
		Timeout:              1 * time.Minute,
		PingFreq:             15 * time.Minute,
		reconnectAttempt:     1,
		QuitMsg:              Version,
	}
	c.AddStdHandlers()
	return c
}

// Connect - see Connection interface docs
func (c *ConnImpl) Connect() error {
	c.quit = false
	c.stopped = true

	if c.Address == nil {
		return ErrInvalidAddress
	} else if len(c.Nick) == 0 {
		return ErrInvalidNick
	} else if len(c.User) == 0 {
		return ErrInvalidUser
	} else if len(c.RealName) == 0 {
		c.RealName = c.User
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

	c.end = make(chan interface{})
	c.output = make(chan *irc.Message, 10)
	c.errors = make(chan error, 2)
	c.disconnected = make(chan error, 2)
	c.Add(3)

	go c.readLoop()
	go c.writeLoop()
	go c.pingLoop()

	for _, auth := range c.Auth {
		auth.Do(c)
	}

	c.SetNick(c.Nick)
	c.SendUser()
	return nil
}

// Loop - see Connection interface docs
func (c *ConnImpl) Loop() {
	for !c.isQuitting() {
		<-c.disconnected
		if c.end != nil {
			close(c.end)
		}
		c.Wait()
		c.Debugfln("Disconnected from server")
		for !c.isQuitting() && !c.Connected() {
			c.Debugln("Trying to reconnect...")
			if err := c.Connect(); err != nil {
				c.Lock()
				nextTime := time.Duration(15*c.reconnectAttempt) * time.Second
				c.Debugfln("Trying again in %v", nextTime)
				if c.reconnectAttempt < 20 {
					c.reconnectAttempt++
				}
				c.Unlock()
				time.Sleep(nextTime)
			} else {
				c.reconnectAttempt = 1
				break
			}
		}
	}
	c.Debugln("Bye!")
}

// LocalAddr - see Connection interface docs
func (c *ConnImpl) LocalAddr() net.Addr {
	return c.socket.LocalAddr()
}

// Disconnect - see Connection interface docs
func (c *ConnImpl) Disconnect() {
	defer recover()
	c.end <- true
	if c.output != nil {
		close(c.output)
	}
	if c.socket != nil {
		c.socket.Close()
	}
	c.Wait()
	c.stopped = true
	c.disconnected <- ErrDisconnected
}

// Connected - see Connection interface docs
func (c *ConnImpl) Connected() bool {
	c.Lock()
	defer c.Unlock()
	return !c.stopped
}

func (c *ConnImpl) isQuitting() bool {
	c.Lock()
	defer c.Unlock()
	return c.quit
}

// GetNick - see Data interface docs
func (c *ConnImpl) GetNick() string {
	return c.Nick
}

// GetPreferredNick - see Data interface docs
func (c *ConnImpl) GetPreferredNick() string {
	return c.PreferredNick
}

// SetQuitMessage - see Data interface docs
func (c *ConnImpl) SetQuitMessage(msg string) {
	c.QuitMsg = msg
}

// SetUseTLS - see Data interface docs
func (c *ConnImpl) SetUseTLS(tls bool) {
	c.UseTLS = tls
}

// SetRealName - see Data interface docs
func (c *ConnImpl) SetRealName(realname string) {
	c.RealName = realname
}

// SetVersion - see Data interface docs
func (c *ConnImpl) SetVersion(version string) {
	c.Version = version
}

// Errors - see ErrorStream interface docs
func (c *ConnImpl) Errors() chan error {
	return c.errors
}

// AddAuth - see Data interface docs
func (c *ConnImpl) AddAuth(auth AuthHandler) {
	c.Auth = append(c.Auth, auth)
}

// SetAddress - see Data interface docs
func (c *ConnImpl) SetAddress(addr Address) {
	c.Address = addr
}

// SetDebugWriter - see Debugger interface docs
func (c *ConnImpl) SetDebugWriter(writer io.Writer) {
	c.DebugWriter = writer
}

// Debugf - see Debugger interface docs
func (c *ConnImpl) Debugf(msg string, args ...interface{}) {
	if c.DebugWriter != nil {
		fmt.Fprintf(c.DebugWriter, msg, args...)
	}
}

// Debugfln - see Debugger interface docs
func (c *ConnImpl) Debugfln(msg string, args ...interface{}) {
	if c.DebugWriter != nil {
		fmt.Fprintf(c.DebugWriter, msg, args...)
		fmt.Fprint(c.DebugWriter, "\n")
	}
}

// Debug - see Debugger interface docs
func (c *ConnImpl) Debug(parts ...interface{}) {
	if c.DebugWriter != nil {
		fmt.Fprint(c.DebugWriter, parts...)
	}
}

// Debugln - see Debugger interface docs
func (c *ConnImpl) Debugln(parts ...interface{}) {
	if c.DebugWriter != nil {
		fmt.Fprintln(c.DebugWriter, parts...)
	}
}
