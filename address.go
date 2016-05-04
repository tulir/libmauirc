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
)

// AddressHandler is an interface with a function that returns a valid connection address.
type AddressHandler interface {
	String() string
}

// IPv4Address implements AddressHandler for IPv4 addresses
type IPv4Address struct {
	IP   string
	Port uint16
}

func (addr IPv4Address) String() string {
	return fmt.Sprintf("%s:%d", addr.IP, addr.Port)
}

// IPv6Address implements AddressHandler for IPv6 addresses
type IPv6Address struct {
	IP   string
	Port uint16
}

func (addr IPv6Address) String() string {
	return fmt.Sprintf("[%s]:%d", addr.IP, addr.Port)
}
