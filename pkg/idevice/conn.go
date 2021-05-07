package idevice

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"io"
	"net"

	"howett.net/plist"
)

type IConn interface {
	Close()
	Reader() io.Reader
	Writer() io.Writer
	Write(data []byte) error
	Encode(msg interface{}) ([]byte, error)
	Decode(r io.Reader) ([]byte, error)
	EnableSessionSSL(cert *Certificate) error
	EnableSessionSSLHandshakeOnly(cert *Certificate) error
}

type Conn struct {
	conn net.Conn
}

func NewConn() (IConn, error) {
	conn, err := net.Dial("unix", "/var/run/usbmuxd")
	if err != nil {
		return nil, err
	}

	return &Conn{conn: conn}, nil
}

func (c *Conn) Close() {
	_ = c.conn.Close()
}

func (c *Conn) Reader() io.Reader {
	return c.conn
}

func (c *Conn) Writer() io.Writer {
	return c.conn
}

func (c *Conn) Write(data []byte) error {
	_, err := c.conn.Write(data)
	return err
}

func (c *Conn) Encode(msg interface{}) ([]byte, error) {
	bs, err := plist.Marshal(msg, plist.XMLFormat)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, uint32(len(bs))); err != nil {
		return nil, err
	}

	buf.Write(bs)
	return buf.Bytes(), nil
}

func (c *Conn) Decode(r io.Reader) ([]byte, error) {
	lenbuf := make([]byte, 4)
	if err := binary.Read(r, binary.BigEndian, lenbuf); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenbuf)
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *Conn) EnableSessionSSLHandshakeOnly(cert *Certificate) error {
	_, err := c.createClientTLSConn(cert)
	if err != nil {
		return err
	}
	return nil
}

func (c *Conn) EnableSessionSSL(cert *Certificate) error {
	conn, err := c.createClientTLSConn(cert)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Conn) createClientTLSConn(cert *Certificate) (*tls.Conn, error) {
	cert5, err := tls.X509KeyPair(cert.HostCertificate, cert.HostPrivateKey)
	if err != nil {
		return nil, err
	}

	tlsConn := tls.Client(c.conn, &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert5},
		ClientAuth:         tls.NoClientCert,
	})
	if err := tlsConn.Handshake(); err != nil {
		return nil, err
	}

	return tlsConn, nil
}
