// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package ntpclient implements NTP request.
//
// Packet format https://tools.ietf.org/html/rfc5905#section-7.3
package ntpclient

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	reserved byte = 0 + iota
	symmetricActive
	symmetricPassive
	client
	server
	broadcast
	controlMessage
	reservedPrivate
)

var (
	errTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	// Timeout is default connection timeout.
	Timeout = time.Duration(5 * time.Second)
	// Port is default NTP server port.
	Port uint = 123
	// Version is default NTP server version.
	Version uint = 4
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

func CustomClient(r Request) (time.Time, error) {
	if r.Version != 4 && r.Version != 3 {
		return errTime, errors.New("invalid version")
	}
	addr := net.JoinHostPort(r.Host, fmt.Sprint(r.Port))
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return errTime, err
	}
	con, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return errTime, err
	}
	defer con.Close()
	con.SetDeadline(time.Now().Add(5 * time.Second))

	data := &msg{}
	// set mode
	// (mode & 11111-000) | 110
	data.LiVnMode = (data.LiVnMode & 0xf8) | client
	// set version
	// (mode & 11-000-111) | xxx-000
	data.LiVnMode = (data.LiVnMode & 0xc7) | byte(r.Version)<<3

	err = binary.Write(con, binary.BigEndian, data)
	if err != nil {
		return errTime, err
	}
	err = binary.Read(con, binary.BigEndian, data)
	if err != nil {
		return errTime, err
	}
	t := data.ReceiveTime.UTC().Local()
	return t, nil
}

// Client send NTP request with default parameters.
func Client(host string) (time.Time, error) {
	r := Request{
		Host:    host,
		Port:    Port,
		Version: Version,
		Timeout: Timeout,
	}
	return CustomClient(r)
}
