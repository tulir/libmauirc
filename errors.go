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
	"errors"
	"fmt"
)

// PreConnError is an error that happened berfore connecting to the server
type PreConnError error

// ConnectionError is an error that happened while connecting to the server.
type ConnectionError struct {
	Cause error
}

func (err ConnectionError) Error() string {
	return fmt.Sprintf("Failed to connect: %v", err.Cause)
}

// Pre-connection errors
var (
	ErrInvalidAddress PreConnError = errors.New("No address given")
	ErrInvalidNick    PreConnError = errors.New("No nick given")
	ErrInvalidUser    PreConnError = errors.New("No user given")
)

// ErrDisconnected is given when the client disconnects
var ErrDisconnected = errors.New("Disconnected")
