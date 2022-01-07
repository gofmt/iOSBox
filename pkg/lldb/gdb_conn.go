package lldb

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
)

type gdbConn struct {
	conn   net.Conn
	rdr    *bufio.Reader
	inbuf  []byte
	outbuf bytes.Buffer

	ack                   bool
	maxTransmitAttrmpts   int
	packetSize            int
	xcmdok                bool
	threadSuffixSupported bool
	multiprocess          bool
}

type gdbRegnames struct {
	PC, SP, BP, CX, FsBase string
}

type GdbProtocolError struct {
	context string
	cmd     string
	code    string
}

func (err *GdbProtocolError) Error() string {
	cmd := err.cmd
	if len(cmd) > 20 {
		cmd = cmd[:20] + "..."
	}
	if err.code == "" {
		return fmt.Sprintf("unsupported packet %s during %s", cmd, err.context)
	}
	return fmt.Sprintf("protocol error %s during %s for packet %s", err.code, err.context, cmd)
}

func isProtocolErrorUnsupported(err error) bool {
	gdberr, ok := err.(*GdbProtocolError)
	if !ok {
		return false
	}
	return gdberr.code == ""
}

const (
	qSupportedSimple       = "$qSupported:swbreak+;hwbreak+;no-resumed+;xmlRegisters=i386"
	qSupportedMultiprocess = "$qSupported:multiprocess+;swbreak+;hwbreak+;no-resumed+;xmlRegisters=i386"
)

func (g *gdbConn) handshake(regnames *gdbRegnames) error {
	g.ack = true
	g.packetSize = 256
	g.rdr = bufio.NewReader(g.conn)

	g.sendack('+')
	g.disableAck()

	if _, err := g.execWithString("$QThreadSuffixSupported"); err != nil {
		if isProtocolErrorUnsupported(err) {
			g.threadSuffixSupported = true
		} else {
			return err
		}
	} else {
		g.threadSuffixSupported = true
	}

	if !g.threadSuffixSupported {
		features, err := g.qSupported(true)
		if err != nil {
			return err
		}
		g.multiprocess = features["multiprocess"]

		if g.multiprocess {
			_, _ = g.execWithString("$Hgp0.0")
		} else {
			_, _ = g.execWithString("$Hgp0")
		}
	} else {
		if _, err := g.qSupported(false); err != nil {
			return err
		}
	}

	// regFound := map[string]bool{
	// 	regnames.PC: false,
	// 	regnames.SP: false,
	// 	regnames.BP: false,
	// 	regnames.CX: false,
	// }

	if resp, err := g.execWithString("$x0,0"); err == nil && string(resp) == "OK" {
		g.xcmdok = true
	}

	return nil
}

func (g *gdbConn) qSupported(multiprocess bool) (features map[string]bool, err error) {
	q := qSupportedSimple
	if multiprocess {
		q = qSupportedMultiprocess
	}
	respBuf, err := g.execWithString(q)
	if err != nil {
		return nil, err
	}
	resp := strings.Split(string(respBuf), ";")
	features = make(map[string]bool)
	for _, stubfeature := range resp {
		if len(stubfeature) <= 0 {
			continue
		} else if equal := strings.Index(stubfeature, "="); equal >= 0 {
			if stubfeature[:equal] == "PacketSize" {
				if n, err := strconv.ParseUint(stubfeature[equal+1:], 16, 64); err != nil {
					g.packetSize = int(n)
				}
			}
		} else if stubfeature[len(stubfeature)-1] == '+' {
			features[stubfeature[:len(stubfeature)-1]] = true
		}
	}

	return
}

func (g *gdbConn) readRegisterInfo(regFound map[string]bool) (err error) {
	// regnum := 0
	// for {
	// 	g.outbuf.Reset()
	// 	_, _ = fmt.Fprintf(&g.outbuf, "$qRegisterInfo%x", regnum)
	// 	respbytes, err := g.exec(g.outbuf.Bytes())
	// 	if err != nil {
	// 		if regnum == 0 {
	// 			return err
	// 		}
	// 		break
	// 	}
	//
	// 	var (
	// 		// regname       string
	// 		// offset        int
	// 		// bitsize       int
	// 		contained     bool
	// 		// ignoreOnWrite bool
	// 	)
	//
	// 	resp := string(respbytes)
	// 	for {
	// 		semicolon := strings.Index(resp, ";")
	// 		keyval := resp
	// 		if semicolon >= 0 {
	// 			keyval = resp[:semicolon]
	// 		}
	//
	// 		colon := strings.Index(keyval, ":")
	// 		if colon >= 0 {
	// 			name := keyval[:colon]
	// 			value := keyval[colon+1:]
	// 			switch name {
	// 			case "name":
	// 				regname = value
	// 			case "offset":
	// 				offset, _ = strconv.Atoi(value)
	// 			case "bitsize":
	// 				bitsize, _ = strconv.Atoi(value)
	// 			case "container-regs":
	// 				contained = true
	// 			case "set":
	// 				if value == "Exception State Registers" {
	// 					ignoreOnWrite = true
	// 				}
	// 			}
	// 		}
	// 		if semicolon < 0 {
	// 			break
	// 		}
	// 		resp = resp[semicolon+1:]
	// 	}
	//
	// 	if contained {
	// 		regnum++
	// 		continue
	// 	}
	//
	// 	regnum++
	// }

	return nil
}

func (g *gdbConn) disableAck() {
	resp, err := g.execWithString("$QStartNoAckMode")
	if err != nil {
		g.ack = true
		return
	}

	g.ack = string(resp) != "OK"
}

type memoryRegionInfo struct {
	start       uint64
	size        uint64
	permissions string
	name        string
}

func decodeHexString(in []byte) (string, bool) {
	out := make([]byte, 0, len(in)/2)
	for i := 0; i < len(in); i += 2 {
		v, err := strconv.ParseUint(string(in[i:i+2]), 16, 8)
		if err != nil {
			return "", false
		}
		out = append(out, byte(v))
	}
	return string(out), true
}

func (g *gdbConn) memoryRegionInfo(addr uint64) (*memoryRegionInfo, error) {
	g.outbuf.Reset()
	_, _ = fmt.Fprintf(&g.outbuf, "$qMemoryRegionInfo:%x", addr)
	resp, err := g.exec(g.outbuf.Bytes())
	if err != nil {
		return nil, err
	}

	mri := &memoryRegionInfo{}
	buf := resp
	for len(buf) > 0 {
		colon := bytes.Index(buf, []byte{':'})
		if colon < 0 {
			break
		}
		key := buf[:colon]
		buf = buf[colon+1:]
		semicolon := bytes.Index(buf, []byte{';'})
		var value []byte
		if semicolon < 0 {
			value = buf
			buf = nil
		} else {
			value = buf[:semicolon]
			buf = buf[semicolon+1:]
		}

		switch string(key) {
		case "start":
			start, err := strconv.ParseUint(string(value), 16, 64)
			if err != nil {
				return nil, xerrors.Errorf("malformed qMemoryRegionInfo response packet (start): %v in %s", err, string(resp))
			}
			mri.start = start
		case "size":
			size, err := strconv.ParseUint(string(value), 16, 64)
			if err != nil {
				return nil, xerrors.Errorf("malformed qMemoryRegionInfo response packet (size): %v in %s", err, string(resp))
			}
			mri.size = size
		case "permissions":
			mri.permissions = string(value)
		case "name":
			namestr, ok := decodeHexString(value)
			if !ok {
				return nil, xerrors.Errorf("malformed qMemoryRegionInfo response packet (name): %s", string(resp))
			}
			mri.name = namestr
		case "error":
			errstr, ok := decodeHexString(value)
			if !ok {
				return nil, xerrors.Errorf("malformed qMemoryRegionInfo response packet (error): %s", string(resp))
			}
			return nil, xerrors.Errorf("qMemoryRegionInfo error: %s", errstr)
		}
	}

	return mri, nil
}

func (g *gdbConn) execWithString(cmd string) ([]byte, error) {
	return g.exec([]byte(cmd))
}

func (g *gdbConn) exec(cmd []byte) ([]byte, error) {
	if err := g.send(cmd); err != nil {
		return nil, err
	}

	return g.recv(cmd)
}

var hexdigit = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

func (g *gdbConn) send(cmd []byte) error {
	if len(cmd) == 0 || cmd[0] != '$' {
		return xerrors.New("command error")
	}

	cmd = append(cmd, '#')
	sum := checksum(cmd)
	cmd = append(cmd, hexdigit[sum>>4], hexdigit[sum&0xf])

	attemt := 0
	for {
		_, err := g.conn.Write(cmd)
		if err != nil {
			return err
		}

		if !g.ack {
			break
		}

		if g.readack() {
			break
		}

		if attemt > g.maxTransmitAttrmpts {
			return xerrors.New("too many transmit attempts")
		}

		attemt++
	}

	return nil
}

func (g *gdbConn) recv(cmd []byte) (resp []byte, err error) {
	attempt := 0
	for {
		resp, err = g.rdr.ReadBytes('#')
		if err != nil {
			return nil, err
		}

		_, err = io.ReadFull(g.rdr, g.inbuf[:2])
		if err != nil {
			return nil, err
		}

		if !g.ack {
			break
		}

		if resp[0] == '%' {
			continue
		}

		if checksumok(resp, g.inbuf[:2]) {
			g.sendack('+')
			break
		}

		if attempt > g.maxTransmitAttrmpts {
			g.sendack('+')
			return nil, xerrors.New("too many transmit attempts")
		}

		attempt++
		g.sendack('-')
	}

	g.inbuf, resp = wiredecode(resp, g.inbuf)
	if len(resp) == 0 || resp[0] == 'E' || (resp[0] == 'E' && len(resp) == 3) {
		cmdstr := ""
		if cmd != nil {
			cmdstr = string(cmd)
		}
		return nil, xerrors.Errorf("%s -> %s", cmdstr, string(resp))
	}

	return resp, nil
}

func (g *gdbConn) readack() bool {
	b, err := g.rdr.ReadByte()
	if err != nil {
		return false
	}

	return b == '+'
}

func (g *gdbConn) sendack(c byte) {
	if c != '+' && c != '-' {
		panic(xerrors.Errorf("sendack(%c)", c))
	}
	_, _ = g.conn.Write([]byte{c})
}

func parsePairValue(data []byte) map[string]string {
	fields := strings.FieldsFunc(string(data), func(r rune) bool {
		return r == ';'
	})

	pairs := make(map[string]string)
	for _, pair := range fields {
		kv := strings.Split(pair, ":")
		pairs[kv[0]] = kv[1]
	}

	return pairs
}

const escapeXor byte = 0x20

func wiredecode(in, buf []byte) (newbuf, msg []byte) {
	if buf != nil {
		buf = buf[:0]
	} else {
		buf = make([]byte, 0, 256)
	}

	start := 1

	for i := 0; i < len(in); i++ {
		switch ch := in[i]; ch {
		case '{': // escape
			if i+1 >= len(in) {
				buf = append(buf, ch)
			} else {
				buf = append(buf, in[i+1]^escapeXor)
				i++
			}
		case ':':
			buf = append(buf, ch)
			if i == 3 {
				// we just read the sequence identifier
				start = i + 1
			}
		case '#': // end of packet
			return buf, buf[start:]
		case '*': // runlength encoding marker
			if i+1 >= len(in) || i == 0 {
				buf = append(buf, ch)
			} else {
				n := in[i+1] - 29
				r := buf[len(buf)-1]
				for j := uint8(0); j < n; j++ {
					buf = append(buf, r)
				}
				i++
			}
		default:
			buf = append(buf, ch)
		}
	}
	return buf, buf[start:]
}

func binarywiredecode(in, buf []byte) (newbuf, msg []byte) {
	if buf != nil {
		buf = buf[:0]
	} else {
		buf = make([]byte, 0, 256)
	}

	start := 1

	for i := 0; i < len(in); i++ {
		switch ch := in[i]; ch {
		case '}': // escape
			if i+1 >= len(in) {
				buf = append(buf, ch)
			} else {
				buf = append(buf, in[i+1]^escapeXor)
				i++
			}
		case '#': // end of packet
			return buf, buf[start:]
		default:
			buf = append(buf, ch)
		}
	}
	return buf, buf[start:]
}

func checksumok(packet, checksumBuf []byte) bool {
	if packet[0] != '$' {
		return false
	}

	sum := checksum(packet)
	tgt, err := strconv.ParseUint(string(checksumBuf), 16, 8)
	if err != nil {
		return false
	}

	return sum == uint8(tgt)
}

func checksum(packet []byte) (sum uint8) {
	for i := 1; i < len(packet); i++ {
		if packet[i] == '#' {
			return sum
		}
		sum += packet[i]
	}
	return
}
