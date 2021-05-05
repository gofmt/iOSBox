package idevice

import (
	"github.com/danielpaulus/go-ios/ios"
	"golang.org/x/xerrors"
	"howett.net/plist"
)

type AppManagerService struct {
	conn       ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

func NewAppManagerService(device ios.DeviceEntry) (*AppManagerService, error) {
	conn, err := ios.ConnectToService(device, "com.apple.mobile.installation_proxy")
	if err != nil {
		return nil, err
	}

	return &AppManagerService{
		conn:       conn,
		plistCodec: ios.NewPlistCodec(),
	}, nil
}

func (a *AppManagerService) Close() {
	a.conn.Close()
}

type AppInstallResponse struct {
	Error           string
	PercentComplete int
	Status          string
}

func (a *AppManagerService) Install(pkgPath string, cb func(AppInstallResponse)) error {
	param := map[string]interface{}{"Command": "Install", "PackagePath": pkgPath}
	bs, err := a.plistCodec.Encode(param)
	if err != nil {
		return err
	}

	if err := a.conn.Send(bs); err != nil {
		return err
	}

	for {
		body, err := a.plistCodec.Decode(a.conn.Reader())
		if err != nil {
			return err
		}

		var resp AppInstallResponse
		if _, err := plist.Unmarshal(body, &resp); err != nil {
			return err
		}

		cb(resp)

		if resp.Error != "" {
			return xerrors.Errorf("%s", resp.Error)
		}

		if resp.Status == "Complete" {
			break
		}
	}

	return nil
}

func (a *AppManagerService) Uninstall(bundleId string) error {
	param := map[string]interface{}{"Command": "Uninstall", "ApplicationIdentifier": bundleId}
	bs, err := a.plistCodec.Encode(param)
	if err != nil {
		return err
	}

	if err := a.conn.Send(bs); err != nil {
		return err
	}

	for {
		body, err := a.plistCodec.Decode(a.conn.Reader())
		if err != nil {
			return err
		}

		var resp AppInstallResponse
		if _, err := plist.Unmarshal(body, &resp); err != nil {
			return err
		}

		if resp.Error != "" {
			return xerrors.Errorf("%s", resp.Error)
		}

		if resp.Status == "Complete" {
			break
		}
	}

	return nil
}
