package idevice

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"howett.net/plist"
)

const (
	TcpdumpMagic     = 0xa1b2c3d4
	PcapVersionMajor = 2
	PcapVersionMinor = 4
	DltEn10mb        = 1
)

type PcapGlobalHeader struct {
	MagicNumber  uint32
	VersionMajor uint16
	VersionMinor uint16
	Thiszone     int32
	Sigfigs      uint32
	Snaplen      uint32
	Network      uint32
}

type PcapPacketHeader struct {
	Timestamp1 uint32
	Timestamp2 uint32
	CapLen     uint32
	Len        uint32
}

type IOSPacketHeader struct {
	HdrLength      uint32
	Version        uint8
	Length         uint32
	Type           uint8
	Unit           uint16
	IO             uint8
	ProtocolFamily uint32
	FramePreLength uint32
	FramePstLength uint32
	IFName         [16]byte
	Pid            uint32
	ProcName       [17]byte
	Unknown        uint32
	Pid2           uint32
	ProcName2      [17]byte
	Unknown2       [8]byte
}

func StartPcapService(ctx context.Context, entry ios.DeviceEntry, procName string, wr io.Writer, dump func([]byte)) error {
	service, err := ios.ConnectToService(entry, "com.apple.pcapd")
	if err != nil {
		return err
	}
	defer service.Close()

	header := PcapGlobalHeader{
		MagicNumber:  TcpdumpMagic,
		VersionMajor: PcapVersionMajor,
		VersionMinor: PcapVersionMinor,
		Thiszone:     0,
		Sigfigs:      0,
		Snaplen:      uint32(65535),
		Network:      uint32(DltEn10mb),
	}

	if err := binary.Write(wr, binary.LittleEndian, header); err != nil {
		return err
	}

	stoped := false
	go func() {
		<-ctx.Done()
		stoped = true
	}()

	pListCodec := ios.NewPlistCodec()
	for {
		if stoped {
			break
		}

		bs, err := pListCodec.Decode(service.Reader())
		if err != nil {
			return err
		}

		_, err = plist.Unmarshal(bs, &bs)
		if err != nil {
			return err
		}

		if dump != nil {
			go dump(bs)
		}

		buf := bytes.NewReader(bs)
		var hdr IOSPacketHeader
		if err := binary.Read(buf, binary.BigEndian, &hdr); err != nil {
			return err
		}

		if len(procName) > 0 {
			pName := string(hdr.ProcName[:])
			pName2 := string(hdr.ProcName2[:])
			if !strings.HasPrefix(pName, procName) && !strings.HasPrefix(pName2, procName) {
				continue
			}
		}

		pphdr := PcapPacketHeader{
			Timestamp1: uint32(time.Now().Unix()),
			Timestamp2: uint32(time.Now().UnixNano() / 1e6),
			CapLen:     hdr.Length,
			Len:        hdr.Length,
		}
		if err := binary.Write(wr, binary.LittleEndian, pphdr); err != nil {
			return err
		}
		if err := binary.Write(wr, binary.LittleEndian, bs[hdr.HdrLength:]); err != nil {
			return err
		}
	}

	return nil
}
