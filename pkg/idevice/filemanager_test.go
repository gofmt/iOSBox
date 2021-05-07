package idevice

import (
	"testing"
)

func TestFileManagerService_FileUpload(t *testing.T) {
	device, err := GetDevice()
	if err != nil {
		t.Fatal(err)
	}

	fileService, err := NewFileManagerService(device)
	if err != nil {
		t.Fatal(err)
	}
	defer fileService.Close()

	handle, err := fileService.FileOpen("./test/test1", AFC_FOPEN_RDONLY)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(handle)

	buf, err := fileService.FileRead(handle, 0x1000)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(buf))

	err = fileService.FileClose(handle)
	if err != nil {
		t.Fatal(err)
	}
}
