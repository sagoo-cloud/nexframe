package guid

import (
	"crypto/rand"
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	sequenceMax   = uint32(46655)
	randomStrBase = "0123456789abcdefghijklmnopqrstuvwxyz"
)

var (
	sequence     uint32 = 0
	macAddrStr          = "0000000"
	processIdStr        = "0000"
)

func init() {
	// MAC addresses hash result in 7 bytes
	ifaces, _ := net.Interfaces()
	var macAddr []byte
	for _, iface := range ifaces {
		if len(iface.HardwareAddr) > 0 {
			macAddr = iface.HardwareAddr
			break
		}
	}
	if macAddr != nil {
		hash := fnv.New32a()
		hash.Write(macAddr)
		s := strconv.FormatUint(uint64(hash.Sum32()), 36)
		copy([]byte(macAddrStr), s)
	}

	// Process id in 4 bytes
	pid := os.Getpid()
	s := strconv.FormatInt(int64(pid), 36)
	copy([]byte(processIdStr), s)
}

func S(data ...[]byte) string {
	b := make([]byte, 32)
	nanoStr := strconv.FormatInt(time.Now().UnixNano(), 36)

	if len(data) == 0 {
		copy(b, macAddrStr)
		copy(b[7:], processIdStr)
		copy(b[11:], nanoStr)
		copy(b[23:], getSequence())
		copy(b[26:], getRandomStr(6))
	} else if len(data) <= 2 {
		n := 0
		for i, v := range data {
			if len(v) > 0 {
				copy(b[i*7:], getDataHashStr(v))
				n += 7
			}
		}
		copy(b[n:], nanoStr)
		copy(b[n+12:], getSequence())
		copy(b[n+12+3:], getRandomStr(32-n-12-3))
	} else {
		panic(fmt.Errorf("too many data parts, it should be no more than 2 parts"))
	}

	return string(b)
}

func getSequence() []byte {
	b := []byte{'0', '0', '0'}
	s := strconv.FormatUint(uint64(atomic.AddUint32(&sequence, 1)%sequenceMax), 36)
	copy(b, s)
	return b
}

func getRandomStr(n int) []byte {
	if n <= 0 {
		return []byte{}
	}
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = randomStrBase[b[i]%36]
	}
	return b
}

func getDataHashStr(data []byte) []byte {
	b := []byte{'0', '0', '0', '0', '0', '0', '0'}
	hash := fnv.New32a()
	hash.Write(data)
	s := strconv.FormatUint(uint64(hash.Sum32()), 36)
	copy(b, s)
	return b
}
