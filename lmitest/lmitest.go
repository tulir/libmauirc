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
package main

import (
	"bufio"
	"flag"
	msg "github.com/sorcix/irc"
	"github.com/sorcix/irc/ctcp"
	irc "maunium.net/go/libmauirc"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var ip = flag.String("address", "localhost", "The address to connect to.")
var port = flag.Int("port", 6667, "The port to connect to.")

func main() {
	flag.Parse()
	c := irc.Create("lmitest", "lmitest", irc.IPv4Address{IP: *ip, Port: uint16(*port)})
	c.RealName = "libmauirc test"
	c.DebugWriter = os.Stdout

	err := c.Connect()
	if err != nil {
		panic(err)
	}

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-term
		c.Debugln("\nInterrupt received...")
		c.Quit()
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()

	go func() {
		err := <-c.Errors
		c.Debugln(err)
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		send := msg.ParseMessage(text)
		if strings.HasPrefix(send.Command, "CTCP_") {
			send.Trailing = ctcp.Encode(send.Command[len("CTCP_"):], send.Trailing)
			send.Command = msg.PRIVMSG
		}
		c.Send(send)
	}
}
