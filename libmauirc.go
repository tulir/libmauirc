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
	"time"
)

// Version is the IRC client version string
var Version = "libmauirc 0.1"

// Connection is an IRC connection
type Connection struct {
	PingFreq      time.Duration
	KeepAlive     time.Duration
	prevMsg       time.Time
	PreferredNick string
	Nick          string
	handlers      map[string][]Handler
}

func (c *Connection) readLoop() {

}

func (c *Connection) writeLoop() {

}

func (c *Connection) pingLoop() {
	//defer c.Done()
	mins := time.NewTicker(1 * time.Minute)
	pingfreq := time.NewTicker(c.PingFreq)
	for {
		select {
		case <-mins.C:
			if time.Since(c.prevMsg) >= c.KeepAlive {
				c.SendRawf("PING %d", time.Now().UnixNano())
			}
		case <-pingfreq.C:
			c.SendRawf("PING %d", time.Now().UnixNano())
			if c.Nick != c.PreferredNick {
				c.Nick = c.PreferredNick
				c.SendRawf("NICK %s", c.Nick)
			}
			//case <-c.end:
			//	ticker.Stop()
			//	ticker2.Stop()
			//	return
		}
	}
}
