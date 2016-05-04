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
	"bufio"
	"github.com/sorcix/irc"
	"strings"
	"time"
)

func (c *Connection) readLoop() {
	defer c.Done()
	br := bufio.NewReaderSize(c.socket, 512)

	for {
		select {
		case <-c.end:
			return
		default:
			if c.socket != nil {
				c.socket.SetReadDeadline(time.Now().Add(c.Timeout + c.PingFreq))
			}

			msg, err := br.ReadString('\n')
			if c.socket != nil {
				var zero time.Time
				c.socket.SetReadDeadline(zero)
			}

			if err != nil {
				c.Errors <- err
				return
			}

			c.Debugln("<--", strings.TrimSpace(msg))

			c.prevMsg = time.Now()
			c.RunHandlers(irc.ParseMessage(msg))
		}
	}
}

func (c *Connection) writeLoop() {
	defer c.Done()
	for {
		select {
		case <-c.end:
			return
		case b, ok := <-c.output:
			if !ok || b == nil || c.socket == nil {
				return
			}

			c.Debugln("-->", strings.TrimSpace(b.String()))
			c.socket.SetWriteDeadline(time.Now().Add(c.Timeout))
			_, err := c.socket.Write(b.Bytes())

			var zero time.Time
			c.socket.SetWriteDeadline(zero)

			if err != nil {
				c.Errors <- err
				return
			}
		}
	}
}

func (c *Connection) pingLoop() {
	defer c.Done()
	mins := time.NewTicker(1 * time.Minute)
	pingfreq := time.NewTicker(c.PingFreq)
	for {
		select {
		case <-mins.C:
			if time.Since(c.prevMsg) >= c.KeepAlive {
				c.Ping()
			}
		case <-pingfreq.C:
			c.Ping()
			if c.Nick != c.PreferredNick {
				c.Nick = c.PreferredNick
				c.SetNick(c.PreferredNick)
			}
		case <-c.end:
			mins.Stop()
			pingfreq.Stop()
			return
		}
	}
}
