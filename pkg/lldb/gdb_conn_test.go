package lldb

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"testing"
)

func TestLLDBProtocol(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		t.Fatal(err)
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	reader := bufio.NewReader(conn)

	_, err = conn.Write([]byte{'+'})
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Write([]byte("$QStartNoAckMode#b0"))
	if err != nil {
		t.Fatal(err)
	}

	plus, err := reader.ReadByte()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(plus))

	buf, err := reader.ReadBytes('#')
	if err != nil {
		t.Fatal(err)
	}

	t.Log(hex.Dump(buf))

	buf = bytes.TrimLeft(buf, "$")
	buf = bytes.TrimRight(buf, "#")

	t.Log(string(buf))

	_, _ = reader.ReadByte()
	_, _ = reader.ReadByte()

	_, err = conn.Write([]byte{'+'})
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Write([]byte("$qHostInfo#00"))
	if err != nil {
		t.Fatal(err)
	}

	buf, err = reader.ReadBytes('#')
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(buf))

	_, _ = reader.ReadByte()
	_, _ = reader.ReadByte()

	_, err = conn.Write([]byte("$qGDBServerVersion#00"))
	if err != nil {
		t.Fatal(err)
	}

	buf, err = reader.ReadBytes('#')
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(buf))

	_, _ = reader.ReadByte()
	_, _ = reader.ReadByte()

	_, err = conn.Write([]byte("$QEnableErrorStrings#00"))
	if err != nil {
		t.Fatal(err)
	}

	buf, err = reader.ReadBytes('#')
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(buf))

	_, _ = reader.ReadByte()
	_, _ = reader.ReadByte()

	binPath := "/var/containers/Bundle/Application/2D28ECCE-BF3E-42BA-BA7A-8E262399C2A8/discover.app/discover"
	_, err = conn.Write([]byte(fmt.Sprintf("$A%d,0,%s#00", len(binPath), binPath)))
	if err != nil {
		t.Fatal(err)
	}

	buf, err = reader.ReadBytes('#')
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(buf))

	_, _ = reader.ReadByte()
	_, _ = reader.ReadByte()
}

func TestGdbCommand(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		t.Fatal(err)
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	gdb := &gdbConn{
		conn:                conn,
		ack:                 true,
		packetSize:          256,
		rdr:                 bufio.NewReader(conn),
		inbuf:               make([]byte, 0, 2048),
		maxTransmitAttrmpts: 3,
	}

	gdb.sendack('+')
	gdb.disableAck()

	resp, err := gdb.execWithString("$qHostInfo")
	if err != nil {
		t.Fatal(err)
	}

	// if string(resp) != "OK" {
	// 	t.Fatal("$QThreadSuffixSupported:", string(resp))
	// }
	//
	// gdb.threadSuffixSupported = true
	//
	// resp, err = gdb.execWithString("$qSupported:swbreak+;hwbreak+;no-resumed+;xmlRegisters=i386")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	t.Log(string(resp))
}

func TestParsePairValue(t *testing.T) {
	ret := parsePairValue([]byte("cputype:16777228;cpusubtype:1;addressing_bits:39;ostype:ios;watchpoint_exceptions_received:before;vendor:apple;os_version:14.0.1;endian:little;ptrsize:8;"))
	t.Log(ret)
}

func TestParseHex(t *testing.T) {
	b, err := hex.DecodeString("2f566f6c756d65732f776f726b2f67636c6179746f6e2f446f63756d656e74732f7372632f6174746163682f612e6f7574")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(b))
}
