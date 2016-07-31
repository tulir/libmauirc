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
	"fmt"
	msg "github.com/sorcix/irc"
	"github.com/sorcix/irc/ctcp"
	irc "maunium.net/go/libmauirc"
	flag "maunium.net/go/mauflag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var ip = flag.Make().ShortKey("a").LongKey("address").Usage("").String()
var port = flag.Make().ShortKey("p").LongKey("port").Usage("").Uint16()
var tls = flag.Make().ShortKey("s").LongKey("ssl").LongKey("tls").Bool()
var wantHelp = flag.Make().ShortKey("h").LongKey("help").Bool()

const help = `lmitest - A simple program to test libmauirc.

Usage:
  lmitest [-s] [-a IP-ADDRESS] [-p PORT]

Help options:
  -h, --help               Show this help page.

Application options:
  -a, --address=IP-ADDRESS The address to connect to.
  -p, --port=PORT          The port to connect to.
  -s, --ssl                Use to enable TLS connection.
`

func main() {
	err := flag.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stdout, help)
		os.Exit(1)
	} else if *wantHelp {
		fmt.Fprintln(os.Stdout, help)
		os.Exit(0)
	}

	c := irc.Create("lmitest", "lmitest", irc.IPv4Address{IP: *ip, Port: *port})
	c.SetRealName("libmauirc tester")
	c.SetDebugWriter(os.Stdout)
	c.SetUseTLS(*tls)

	err = c.Connect()
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
		err := <-c.Errors()
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
