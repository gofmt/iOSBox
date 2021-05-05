package idevice

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
)

func TestAppManagerService_Install(t *testing.T) {
	deviceList, err := ios.ListDevices()
	if err != nil {
		t.Fatal(err)
	}

	if len(deviceList.DeviceList) == 0 {
		t.Fatal("没有IOS设备")
	}

	device := deviceList.DeviceList[0]

	fileService, err := NewFileManagerService(device)
	if err != nil {
		t.Fatal(err)
	}
	defer fileService.Close()

	ipaPath := "../../testdata/蜂鸟众包_7.10.1.ipa"
	remotePath := "PublicStaging/" + filepath.Base(ipaPath)
	lfile, err := os.Open(ipaPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(lfile)
	if err := fileService.FileUpload(lfile, remotePath, func(count int) {
		t.Log(count)
	}); err != nil {
		t.Fatal(err)
	}

	appService, err := NewAppManagerService(device)
	if err != nil {
		t.Fatal(err)
	}
	defer appService.Close()

	if err := appService.Install(remotePath, func(ret AppInstallResponse) {
		t.Log(ret)
	}); err != nil {
		t.Fatal(err)
	}
}
