// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package ntpclient implements NTP request.
package ntpclient

import (
	// "encoding/binary"
	"net"
	"time"
)

type Request struct {
	Host    string
	Port    uint
	Version uint
	Timeout time.Duration
}

type ntpTime struct {
	Seconds  uint32
	Fraction uint32
}

type msg struct {
	LiVnMode       byte // Leap Indicator (2) + Version (3) + Mode (3)
	Stratum        byte
	Poll           byte
	Precision      byte
	RootDelay      uint32
	RootDispersion uint32
	ReferenceId    uint32
	ReferenceTime  ntpTime
	OriginTime     ntpTime
	ReceiveTime    ntpTime
	TransmitTime   ntpTime
}

func (t ntpTime) UTC() time.Time {
	nsec := uint64(t.Seconds)*1e9 + (uint64(t.Fraction) * 1e9 >> 32)
	return time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(nsec))
}

func send(r *Request) (time.Time, error) {
	// validate host/port
	// set version
	// set net deadline
	return time.Now(), error
}

func CustomClient(r Request) (time.Time, error) {
	return time.Now(), error
}

func Client(host string) (time.Time, error) {
	return time.Now(), error
}
