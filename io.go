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
	"bytes"
	"strings"
	"time"

	"github.com/sorcix/irc"
)

func (c *ConnImpl) readLoop() {
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
				c.errors <- err
				return
			}

			msg = strings.TrimSpace(msg)
			c.Debugln("<--", msg)
			c.Lock()
			c.prevMsg = time.Now()
			c.Unlock()
			if strings.HasPrefix(msg, "ERROR") {
				go c.Disconnect()
				return
			}
			c.RunHandlers(irc.ParseMessage(msg))
		}
	}
}

func (c *ConnImpl) writeLoop() {
	defer c.Done()
	for {
		select {
		case b, ok := <-c.output:
			if !ok || b == nil || c.socket == nil {
				return
			}

			c.Debugln("-->", strings.TrimSpace(b.String()))
			c.socket.SetWriteDeadline(time.Now().Add(c.Timeout))
			var buf bytes.Buffer
			buf.Write(b.Bytes())
			buf.WriteRune('\r')
			buf.WriteRune('\n')
			_, err := c.socket.Write(buf.Bytes())

			var zero time.Time
			c.socket.SetWriteDeadline(zero)

			if err != nil {
				c.errors <- err
				return
			}
		case <-c.end:
			return
		}
	}
}

func (c *ConnImpl) pingLoop() {
	defer c.Done()
	mins := time.NewTicker(1 * time.Minute)
	pingfreq := time.NewTicker(c.PingFreq)
	defer mins.Stop()
	defer pingfreq.Stop()
	for {
		select {
		case <-mins.C:
			c.Lock()
			if time.Since(c.prevMsg) >= c.KeepAlive {
				c.Ping()
			}
			c.Unlock()
		case <-pingfreq.C:
			c.Ping()
			c.Lock()
			if c.Nick != c.PreferredNick {
				c.Nick = c.PreferredNick
				c.SetNick(c.PreferredNick)
			}
			c.Unlock()
		case <-c.end:
			return
		}
	}
}
