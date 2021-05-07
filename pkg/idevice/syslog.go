package idevice

import (
	"bufio"
	"bytes"
	"strings"
)

type LogMessage struct {
	Time       string
	DeviceName string
	ProcInfo   string
	Level      string
	Body       string
}

type SyslogService struct {
	conn   IConn
	br     *bufio.Reader
	closed bool
}

func NewSyslogService(entry *DeviceEntry) (*SyslogService, error) {
	conn, err := ConnectToService(entry, "com.apple.syslog_relay")
	if err != nil {
		return &SyslogService{}, err
	}

	return &SyslogService{
		conn:   conn,
		br:     bufio.NewReader(conn.Reader()),
		closed: false,
	}, nil
}

func (s *SyslogService) Close() {
	s.closed = true
	s.conn.Close()
}

func (s *SyslogService) GetSyslog() (LogMessage, error) {
	bs, err := s.br.ReadBytes(0)
	if err != nil {
		if s.closed {
			return LogMessage{}, nil
		}
		return LogMessage{}, err
	}

	bs = bytes.TrimRight(bs, "\x0a\x00")
	line := decodeSyslog(bs)
	ss := strings.Split(line, ">: ")
	header := strings.Split(ss[0], " ")

	return LogMessage{
		Time:       header[0] + " " + header[2] + " " + header[3],
		DeviceName: header[4],
		ProcInfo:   header[5],
		Level:      header[6][1:],
		Body:       ss[1],
	}, nil
}

func decodeSyslog(bs []byte) string {
	specialChar := bytes.Contains(bs, []byte(`\134`))
	if specialChar {
		bs = bytes.Replace(bs, []byte(`\134`), []byte(""), -1)
	}
	kBackslash := byte(0x5c)
	kM := byte(0x4d)
	kDash := byte(0x2d)
	kCaret := byte(0x5e)

	// Mask for the UTF-8 digit range.
	kNum := byte(0x30)

	var out []byte
	for i := 0; i < len(bs); {

		if (bs[i] != kBackslash) || i > (len(bs)-4) {
			out = append(out, bs[i])
			i = i + 1
		} else {
			if (bs[i+1] == kM) && (bs[i+2] == kCaret) {
				out = append(out, (bs[i+3]&byte(0x7f))+byte(0x40))
			} else if bs[i+1] == kM && bs[i+2] == kDash {
				out = append(out, bs[i+3]|byte(0x80))
			} else if isDigit(bs[i+1:i+3], kNum) {
				out = append(out, decodeOctal(bs[i+1], bs[i+2], bs[i+3]))
			} else {
				out = append(out, bs[0], bs[1], bs[2], bs[3], bs[4])
			}
			i = i + 4
		}
	}
	return string(out)
}

func isDigit(b []byte, kNum byte) bool {
	for _, v := range b {
		if (v & byte(0xf0)) != kNum {
			return false
		}
	}
	return true
}

func decodeOctal(x, y, z byte) byte {
	return (x&byte(0x3))<<byte(6) | (y&byte(0x7))<<byte(3) | z&byte(0x7)
}
