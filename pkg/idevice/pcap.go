package idevice

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"

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
	SubPid         uint32
	SubProcName    [17]byte
	Unknown2       [8]byte
}

func StartPcapService(ctx context.Context, entry *DeviceEntry, procName string, wr io.Writer, dump func([]byte)) error {
	service, err := ConnectToService(entry, "com.apple.pcapd")
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
		fmt.Println("exit")
	}()

	for {
		bs, err := service.Decode(service.Reader())
		if err != nil {
			return err
		}

		var data []byte
		_, err = plist.Unmarshal(bs, &data)
		if err != nil {
			return err
		}

		buf := bytes.NewReader(data)
		var hdr IOSPacketHeader
		if err := binary.Read(buf, binary.BigEndian, &hdr); err != nil {
			return err
		}

		if len(procName) > 0 {
			pName := strings.TrimSpace(string(hdr.ProcName[:]))
			subName := strings.TrimSpace(string(hdr.SubProcName[:]))
			if !strings.HasPrefix(pName, procName) && !strings.HasPrefix(subName, procName) {
				continue
			}
		}

		if dump != nil {
			go dump(data)
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

		if hdr.FramePreLength == 0 {
			fmt.Printf("%+v\n", hdr)
			ext := []byte{0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0x08, 0x00}
			body := append(ext, data[hdr.HdrLength:]...)
			err = binary.Write(wr, binary.LittleEndian, body)
		} else {
			err = binary.Write(wr, binary.LittleEndian, data[hdr.HdrLength:])
		}

		if err != nil {
			return err
		}

		if stoped {
			break
		}
	}

	return nil
}
