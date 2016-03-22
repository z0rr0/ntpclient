// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ntpclient implements NTP request.
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

// Request contains main NTP request parameters.
type Request struct {
	Host    string
	Port    uint
	Version uint
	Timeout time.Duration
}

// Response is short NTP client response
type Response struct {
	L      time.Time     // local
	R      time.Time     // remote
	Diff   time.Duration // time delta
	Statum int           // stratum value
	Err    error         // error flag
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
	ReferenceID    uint32
	ReferenceTime  ntpTime
	OriginTime     ntpTime
	ReceiveTime    ntpTime
	TransmitTime   ntpTime
}

// UTC returns NTP client UTC time.
func (t ntpTime) UTC() time.Time {
	nsec := uint64(t.Seconds)*1e9 + (uint64(t.Fraction) * 1e9 >> 32)
	return time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(nsec))
}

func (m msg) diff(original time.Time) time.Duration {
	original2 := time.Now()
	duration := original2.Sub(original)
	processing := m.TransmitTime.UTC().Sub(m.ReceiveTime.UTC())

	sendTime := (duration - processing) / 2
	delta := m.ReceiveTime.UTC().Sub(original) - sendTime

	// sendNsec := uint64(m.RootDelay>>16)*1e9 + ((uint64(m.RootDelay&0x0000ffff) * 1e9) >> 16)
	return delta
}

// get writes and read UDP socket data,
// also it calculates and returns initial local time.
func get(m *msg, con *net.UDPConn, version uint) (time.Time, error) {
	// (mode & 11111-000) | 110
	m.LiVnMode = (m.LiVnMode & 0xf8) | client
	// set version
	// (mode & 11-000-111) | xxx-000
	m.LiVnMode = (m.LiVnMode & 0xc7) | byte(version)<<3
	original := time.Now()
	err := binary.Write(con, binary.BigEndian, m)
	if err != nil {
		return original, err
	}
	return original, binary.Read(con, binary.BigEndian, m)
}

// CustomClient is a custom NTP request.
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
	// send/get NTP data
	data := &msg{}
	_, err = get(data, con, r.Version)
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

// ExtClient is extended NTP client.
func ExtClient(r Request) Response {
	if r.Version != 4 && r.Version != 3 {
		return Response{Err: errors.New("invalid version")}
	}
	addr := net.JoinHostPort(r.Host, fmt.Sprint(r.Port))
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return Response{Err: err}
	}
	con, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return Response{Err: err}
	}
	defer con.Close()
	con.SetDeadline(time.Now().Add(5 * time.Second))
	// send/get NTP data
	data := &msg{}
	original, err := get(data, con, r.Version)
	if err != nil {
		return Response{Err: err}
	}
	// tmp := uint64(data.RootDelay>>16)*1e9 + ((uint64(data.RootDelay&0x0000ffff) * 1e9) >> 16)
	// fmt.Println(tmp)

	return Response{
		Diff:   data.diff(original),
		L:      original,
		R:      data.ReceiveTime.UTC().Local(),
		Statum: int(data.Stratum),
	}
}
