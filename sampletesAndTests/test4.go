package main

import (
	"encoding/binary"
	"fmt"
	"time"
)

type t struct {
	tm time.Time
}

func bin2time(sec uint64, nsec uint32) time.Time {
	t := time.Now()
	var bintime [1 + 8 + 4 + 2]byte
	bintime[0] = 1
	binary.BigEndian.PutUint64(bintime[1:], sec)
	binary.BigEndian.PutUint32(bintime[9:], nsec)
	fmt.Printf("% x\n", bintime)

	t.UnmarshalBinary(bintime[:])
	return t
}

func time2bin(t time.Time) (sec uint64, nsec uint32) {
	b, _ := t.MarshalBinary()
	fmt.Printf("% x\n", b)

	sec = binary.BigEndian.Uint64(b[1:9])
	nsec = binary.BigEndian.Uint32(b[9:13])
	return
}

func main() {
	now := time.Now()

	// show internal time struct
	fmt.Printf("%s, internal %v\n", now, t{now})

	// convert the time to binary, then get internal vars from there
	sec, nsec := time2bin(now)
	fmt.Printf("s %d, ns %d\n", sec, nsec)

	// convert the internal vars back to a time
	t2 := bin2time(sec, nsec+1000)
	fmt.Printf("%s, internal %v\n", t2, t{t2})
}
