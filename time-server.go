//
// Flexible time server (RFC 868)
// http://www.faqs.org/rfcs/rfc868.html
//
//

package main

import (
	"os"
	"fmt"
	"strings"
	"time"
	"net"
	"io/ioutil"
	"encoding/binary" )

const (
	Add = "+"
	Subtract = "-"
	Fix = "fix"
	TIME_OFFSET int64 = 2208988800 // time at 00:00:00 1 Jan 1970 GMT
)

// Op can be "+", "-" or "fix" (case sensitive)
type Offset struct {
	Op string
	Value int
}

// per-IP offsets, read from file
type HostOffset struct {
	Year, Month, Day, Hour, Minute, Second Offset
}

// Read in the offsets for an IP from a file. File name == IP address.
func GetOffset(addr string) (HostOffset, os.Error) {
	data, err := ioutil.ReadFile(addr)
	o := HostOffset{}

	if err != nil {
		return o, err
	}

	str := string(data)

	_, err = fmt.Sscanln(str,
		&o.Year.Op,	&o.Year.Value,
		&o.Month.Op,	&o.Month.Value,
		&o.Day.Op,	&o.Day.Value,
		&o.Hour.Op,	&o.Hour.Value,
		&o.Minute.Op,	&o.Minute.Value,
		&o.Second.Op,	&o.Second.Value)

	return o, err
}

// Apply offset to local time.
// Returns a Time object: http://golang.org/pkg/time/#Time
func GetFixedUpTime(o HostOffset) (time.Time) {
	tm := *time.LocalTime()
	
	switch o.Year.Op {
	case Add:	tm.Year += int64(o.Year.Value)
	case Subtract:	tm.Year -= int64(o.Year.Value)
	case Fix:	tm.Year = int64(o.Year.Value)
	}

	switch o.Month.Op {
	case Add:	tm.Month += o.Month.Value
	case Subtract:	tm.Month -= o.Month.Value
	case Fix:	tm.Month = o.Month.Value
	}

	switch o.Day.Op {
	case Add:	tm.Day += o.Day.Value
	case Subtract:	tm.Day -= o.Day.Value
	case Fix:	tm.Day = o.Day.Value
	}

	switch o.Hour.Op {
	case Add:	tm.Hour += o.Hour.Value
	case Subtract:	tm.Hour -= o.Hour.Value
	case Fix:	tm.Hour = o.Hour.Value
	}

	switch o.Minute.Op {
	case Add:	tm.Minute += o.Minute.Value
	case Subtract:	tm.Minute -= o.Minute.Value
	case Fix:	tm.Minute = o.Minute.Value
	}

	switch o.Second.Op {
	case Add:	tm.Second += o.Second.Value
	case Subtract:	tm.Second -= o.Second.Value
	case Fix:	tm.Second = o.Second.Value
	}

	// fmt.Println(tm.String())
	return tm
}

//
// See also FreeBSD inetd implementation:
// http://www.freebsd.org/cgi/cvsweb.cgi/src/usr.sbin/inetd/builtins.c?rev=1.45.14.1;content-type=text%2Fplain
// machtime()
//
func RFC868Time(tm time.Time) uint32 {
	tm64 := tm.Seconds()
	// fmt.Println(tm64)
	return uint32((tm64 + TIME_OFFSET) & 0xFFFFFFFF)
}

func ServeTime(conn net.Conn) {
	defer conn.Close()

	remaddr := conn.RemoteAddr().String()
	remIP := strings.Split(remaddr, ":")[0]
	fmt.Println("Accepted connection from " + remIP)

	hostOff, err := GetOffset(remIP)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get offset (" + err.String() + "), closing connection")
		return
	}

	// tm32 := uint32((time.Seconds() + TIME_OFFSET) & 0xFFFFFFFF)
	tm32 := RFC868Time(GetFixedUpTime(hostOff))
	// fmt.Println(time.LocalTime().Seconds())
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, tm32)
	conn.Write(buf)
}

func RDateListen() {
	listener, err := net.Listen("tcp4", "0.0.0.0:37")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for true {
		conn, aerr := listener.Accept()
		if aerr != nil {
			fmt.Fprintln(os.Stderr, aerr)
			continue
		}
		go ServeTime(conn)
	}
}

func main() {
	fmt.Println("Time server starting")
	RDateListen()
}
