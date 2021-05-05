package idevice

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/danielpaulus/go-ios/ios"
	"golang.org/x/xerrors"
)

const (
	AFC_OP_INVALID                   = 0x00000000 /* Invalid */
	AFC_OP_STATUS                    = 0x00000001 /* Status */
	AFC_OP_DATA                      = 0x00000002 /* Data */
	AFC_OP_READ_DIR                  = 0x00000003 /* ReadDir */
	AFC_OP_READ_FILE                 = 0x00000004 /* ReadFile */
	AFC_OP_WRITE_FILE                = 0x00000005 /* WriteFile */
	AFC_OP_WRITE_PART                = 0x00000006 /* WritePart */
	AFC_OP_TRUNCATE                  = 0x00000007 /* TruncateFile */
	AFC_OP_REMOVE_PATH               = 0x00000008 /* RemovePath */
	AFC_OP_MAKE_DIR                  = 0x00000009 /* MakeDir */
	AFC_OP_GET_FILE_INFO             = 0x0000000A /* GetFileInfo */
	AFC_OP_GET_DEVINFO               = 0x0000000B /* GetDeviceInfo */
	AFC_OP_WRITE_FILE_ATOM           = 0x0000000C /* WriteFileAtomic (tmp file+rename) */
	AFC_OP_FILE_OPEN                 = 0x0000000D /* FileRefOpen */
	AFC_OP_FILE_OPEN_RES             = 0x0000000E /* FileRefOpenResult */
	AFC_OP_FILE_READ                 = 0x0000000F /* FileRefRead */
	AFC_OP_FILE_WRITE                = 0x00000010 /* FileRefWrite */
	AFC_OP_FILE_SEEK                 = 0x00000011 /* FileRefSeek */
	AFC_OP_FILE_TELL                 = 0x00000012 /* FileRefTell */
	AFC_OP_FILE_TELL_RES             = 0x00000013 /* FileRefTellResult */
	AFC_OP_FILE_CLOSE                = 0x00000014 /* FileRefClose */
	AFC_OP_FILE_SET_SIZE             = 0x00000015 /* FileRefSetFileSize (ftruncate) */
	AFC_OP_GET_CON_INFO              = 0x00000016 /* GetConnectionInfo */
	AFC_OP_SET_CON_OPTIONS           = 0x00000017 /* SetConnectionOptions */
	AFC_OP_RENAME_PATH               = 0x00000018 /* RenamePath */
	AFC_OP_SET_FS_BS                 = 0x00000019 /* SetFSBlockSize (0x800000) */
	AFC_OP_SET_SOCKET_BS             = 0x0000001A /* SetSocketBlockSize (0x800000) */
	AFC_OP_FILE_LOCK                 = 0x0000001B /* FileRefLock */
	AFC_OP_MAKE_LINK                 = 0x0000001C /* MakeLink */
	AFC_OP_GET_FILE_HASH             = 0x0000001D /* GetFileHash */
	AFC_OP_SET_FILE_MOD_TIME         = 0x0000001E /* SetModTime */
	AFC_OP_GET_FILE_HASH_RANGE       = 0x0000001F /* GetFileHashWithRange */
	AFC_OP_FILE_SET_IMMUTABLE_HINT   = 0x00000020 /* FileRefSetImmutableHint */
	AFC_OP_GET_SIZE_OF_PATH_CONTENTS = 0x00000021 /* GetSizeOfPathContents */
	AFC_OP_REMOVE_PATH_AND_CONTENTS  = 0x00000022 /* RemovePathAndContents */
	AFC_OP_DIR_OPEN                  = 0x00000023 /* DirectoryEnumeratorRefOpen */
	AFC_OP_DIR_OPEN_RESULT           = 0x00000024 /* DirectoryEnumeratorRefOpenResult */
	AFC_OP_DIR_READ                  = 0x00000025 /* DirectoryEnumeratorRefRead */
	AFC_OP_DIR_CLOSE                 = 0x00000026 /* DirectoryEnumeratorRefClose */
	AFC_OP_FILE_READ_OFFSET          = 0x00000027 /* FileRefReadWithOffset */
	AFC_OP_FILE_WRITE_OFFSET         = 0x00000028 /* FileRefWriteWithOffset */
)

type FileMode uint64

const (
	AFC_FOPEN_RDONLY   FileMode = 0x00000001 /**< r   O_RDONLY */
	AFC_FOPEN_RW       FileMode = 0x00000002 /**< r+  O_RDWR   | O_CREAT */
	AFC_FOPEN_WRONLY   FileMode = 0x00000003 /**< w   O_WRONLY | O_CREAT  | O_TRUNC */
	AFC_FOPEN_WR       FileMode = 0x00000004 /**< w+  O_RDWR   | O_CREAT  | O_TRUNC */
	AFC_FOPEN_APPEND   FileMode = 0x00000005 /**< a   O_WRONLY | O_APPEND | O_CREAT */
	AFC_FOPEN_RDAPPEND FileMode = 0x00000006 /**< a+  O_RDWR   | O_APPEND | O_CREAT */
)

var DefaultChunkSize = 1048576

type AFCHeader struct {
	Magic        [8]byte
	EntireLength uint64
	ThisLength   uint64
	PacketNum    uint64
	Operation    uint64
}

type AFCPacket struct {
	header  AFCHeader
	param   []byte
	payload []byte
}

type MapResult map[string]interface{}

type FileManagerService struct {
	conn       ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
	header     *AFCHeader
}

func NewFileManagerService(device ios.DeviceEntry) (*FileManagerService, error) {
	conn, err := ios.ConnectToService(device, "com.apple.afc")
	if err != nil {
		return nil, err
	}

	var magic [8]byte
	copy(magic[:], "CFA6LPAA")

	return &FileManagerService{
		conn:       conn,
		plistCodec: ios.NewPlistCodec(),
		header: &AFCHeader{
			Magic:        magic,
			EntireLength: 0,
			ThisLength:   0,
			PacketNum:    0,
			Operation:    0,
		},
	}, nil
}

func (f *FileManagerService) Close() {
	f.conn.Close()
}

func (f *FileManagerService) GetDeviceInfo() (MapResult, error) {
	if err := f.Send(AFC_OP_GET_DEVINFO, nil, nil); err != nil {
		return nil, err
	}

	ret, err := f.Recv()
	if err != nil {
		return nil, err
	}

	return f.buildMapResult(ret.param), nil
}

func (f *FileManagerService) ReadDir(dir string) ([]string, error) {
	if err := f.Send(AFC_OP_READ_DIR, []byte(dir), nil); err != nil {
		return nil, err
	}

	ret, err := f.Recv()
	if err != nil {
		return nil, err
	}

	return f.buildSliceResult(ret.payload), nil
}

func (f *FileManagerService) MakeDir(dir string) (uint64, error) {
	if err := f.Send(AFC_OP_MAKE_DIR, []byte(dir), nil); err != nil {
		return 0, err
	}

	ret, err := f.Recv()
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(ret.param), nil
}

func (f *FileManagerService) RemovePath(path string) (uint64, error) {
	if err := f.Send(AFC_OP_REMOVE_PATH, []byte(path), nil); err != nil {
		return 0, err
	}

	ret, err := f.Recv()
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(ret.param), nil
}

func (f *FileManagerService) GetFileInfo(path string) (MapResult, error) {
	if err := f.Send(AFC_OP_GET_FILE_INFO, []byte(path), nil); err != nil {
		return nil, err
	}

	ret, err := f.Recv()
	if err != nil {
		return nil, err
	}

	return f.buildMapResult(ret.payload), nil
}

// FileOpen /var/mobile/Media
func (f *FileManagerService) FileOpen(fileName string, mode FileMode) (uint64, error) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(mode))
	param := append(buf, []byte(fileName)...)
	param = append(param, []byte{0x00}...)
	if err := f.Send(AFC_OP_FILE_OPEN, param, nil); err != nil {
		return 0, err
	}

	ret, err := f.Recv()
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(ret.param), nil
}

func (f *FileManagerService) FileClose(handle uint64) error {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, handle)
	if err := f.Send(AFC_OP_FILE_CLOSE, data, nil); err != nil {
		return err
	}

	ret, err := f.Recv()
	if err != nil {
		return err
	}

	if binary.LittleEndian.Uint64(ret.param) != 0 {
		return xerrors.New("file close error")
	}

	return nil
}

func (f *FileManagerService) FileWrite(handle uint64, data []byte) error {
	bHandle := make([]byte, 8)
	binary.LittleEndian.PutUint64(bHandle, handle)
	if err := f.Send(AFC_OP_FILE_WRITE, bHandle, data); err != nil {
		return err
	}

	ret, err := f.Recv()
	if err != nil {
		return err
	}

	if binary.LittleEndian.Uint64(ret.param) != 0 {
		return xerrors.New("file write error")
	}

	return nil
}

func (f *FileManagerService) FileRead(handle, size uint64) ([]byte, error) {
	buf := make([]byte, 0)

	bHandle := make([]byte, 8)
	binary.LittleEndian.PutUint64(bHandle, handle)
	buf = append(buf, bHandle...)

	bHandle = make([]byte, 8)
	binary.LittleEndian.PutUint64(bHandle, size)
	buf = append(buf, bHandle...)
	if err := f.Send(AFC_OP_FILE_READ, buf, nil); err != nil {
		return nil, err
	}

	ret, err := f.Recv()
	if err != nil {
		return nil, err
	}

	return ret.payload, nil
}

func (f *FileManagerService) FileUpload(local io.Reader, remote string, cb func(int)) error {
	handle, err := f.FileOpen(remote, AFC_FOPEN_WRONLY)
	if err != nil {
		return err
	}
	defer func(f *FileManagerService, handle uint64) {
		_ = f.FileClose(handle)
	}(f, handle)

	buf := make([]byte, DefaultChunkSize)
	amount := 0
	for {
		switch nr, err := local.Read(buf[:]); true {
		case nr < 0:
			return xerrors.Errorf("cat: error reading: %w", err)
		case nr == 0:
			return nil
		case nr > 0:
			cb(amount)
			if err := f.FileWrite(handle, buf); err != nil {
				return err
			}
			amount++
		}
	}
}

func (f *FileManagerService) Send(op int, param, payload []byte) error {
	paramLen := len(param)
	payloadLen := len(payload)

	f.header.PacketNum++
	f.header.EntireLength = uint64(40 + paramLen + payloadLen)
	f.header.ThisLength = uint64(40 + paramLen)
	f.header.Operation = uint64(op)

	// fmt.Printf("req: %#v\n", f.header)

	wr := f.conn.Writer()
	if err := binary.Write(wr, binary.LittleEndian, f.header); err != nil {
		return err
	}

	if paramLen > 0 {
		if err := binary.Write(wr, binary.LittleEndian, param); err != nil {
			return err
		}
	}

	if payloadLen > 0 {
		if err := binary.Write(wr, binary.LittleEndian, payload); err != nil {
			return err
		}
	}

	return nil
}

func (f *FileManagerService) Recv() (AFCPacket, error) {
	rr := f.conn.Reader()

	var header AFCHeader
	if err := binary.Read(rr, binary.LittleEndian, &header); err != nil {
		return AFCPacket{}, err
	}

	if string(header.Magic[:]) != "CFA6LPAA" {
		return AFCPacket{}, xerrors.Errorf("Invalid AFC packet received")
	}

	if header.PacketNum != f.header.PacketNum {
		return AFCPacket{}, xerrors.Errorf("Unexpected packet number (%d != %d) aborting.",
			header.PacketNum, f.header.PacketNum)
	}

	// fmt.Printf("resp: %#v\n", header)

	param := make([]byte, header.ThisLength-40)
	if _, err := io.ReadFull(rr, param); err != nil {
		return AFCPacket{}, err
	}

	payload := make([]byte, header.EntireLength-header.ThisLength)
	if _, err := io.ReadFull(rr, payload); err != nil {
		return AFCPacket{}, err
	}

	return AFCPacket{
		header:  header,
		param:   param,
		payload: payload,
	}, nil
}

func (f *FileManagerService) buildSliceResult(buf []byte) []string {
	result := make([]string, 0)
	bs := bytes.Split(buf, []byte{0x00})
	for _, b := range bs {
		result = append(result, string(b))
	}

	return result
}

func (f *FileManagerService) buildMapResult(buf []byte) MapResult {
	m := make(MapResult)
	bs := bytes.Split(buf, []byte{0x00})
	for i := 0; i < len(bs); i += 2 {
		if i == len(bs)-1 {
			break
		}
		m[string(bs[i])] = string(bs[i+1])
	}

	return m
}
