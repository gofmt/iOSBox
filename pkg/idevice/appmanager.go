package idevice

import (
	"golang.org/x/xerrors"
	"howett.net/plist"
)

type BrowseResponse struct {
	CurrentIndex  uint64
	CurrentAmount uint64
	Status        string
	CurrentList   []AppInfo
}
type AppInfo struct {
	ApplicationDSID              int
	ApplicationType              string
	CFBundleDisplayName          string
	CFBundleExecutable           string
	CFBundleIdentifier           string
	CFBundleName                 string
	CFBundleShortVersionString   string
	CFBundleVersion              string
	Container                    string
	Entitlements                 map[string]interface{}
	EnvironmentVariables         map[string]interface{}
	MinimumOSVersion             string
	Path                         string
	ProfileValidated             bool
	SBAppTags                    []string
	SignerIdentity               string
	UIDeviceFamily               []int
	UIRequiredDeviceCapabilities []string
}

type AppManagerService struct {
	conn IConn
}

func NewAppManagerService(device *DeviceEntry) (*AppManagerService, error) {
	conn, err := ConnectToService(device, "com.apple.mobile.installation_proxy")
	if err != nil {
		return nil, err
	}

	return &AppManagerService{
		conn: conn,
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

func (a *AppManagerService) GetApplications() ([]AppInfo, error) {
	clientOptions := map[string]interface{}{
		"ApplicationType": "User",
		"ReturnAttributes": []string{
			"ApplicationDSID",
			"ApplicationType",
			"CFBundleDisplayName",
			"CFBundleExecutable",
			"CFBundleIdentifier",
			"CFBundleName",
			"CFBundleShortVersionString",
			"CFBundleVersion",
			"Container",
			"Entitlements",
			"EnvironmentVariables",
			"MinimumOSVersion",
			"Path",
			"ProfileValidated",
			"SBAppTags",
			"SignerIdentity",
			"UIDeviceFamily",
			"UIRequiredDeviceCapabilities",
		},
	}

	param := map[string]interface{}{"ClientOptions": clientOptions, "Command": "Browse"}
	userApps, err := a.browseApps(param)
	if err != nil {
		return nil, err
	}

	clientOptions["ApplicationType"] = "System"
	sysApps, err := a.browseApps(param)
	if err != nil {
		return nil, err
	}

	return append(userApps, sysApps...), nil
}

func (a *AppManagerService) browseApps(param map[string]interface{}) ([]AppInfo, error) {
	bs, err := a.conn.Encode(param)
	if err != nil {
		return nil, err
	}

	if err := a.conn.Write(bs); err != nil {
		return nil, err
	}

	resps := make([]BrowseResponse, 0)
	next := true
	count := uint64(0)
	for next {
		body, err := a.conn.Decode(a.conn.Reader())
		if err != nil {
			return nil, err
		}

		var resp BrowseResponse
		if _, err := plist.Unmarshal(body, &resp); err != nil {
			return nil, err
		}

		next = resp.Status != "Complete"
		count += resp.CurrentAmount
		resps = append(resps, resp)
	}

	apps := make([]AppInfo, count)
	for _, resp := range resps {
		copy(apps[resp.CurrentIndex:], resp.CurrentList)
	}

	return apps, nil
}

func (a *AppManagerService) Install(pkgPath string, cb func(AppInstallResponse)) error {
	param := map[string]interface{}{"Command": "Install", "PackagePath": pkgPath}
	bs, err := a.conn.Encode(param)
	if err != nil {
		return err
	}

	if err := a.conn.Write(bs); err != nil {
		return err
	}

	for {
		body, err := a.conn.Decode(a.conn.Reader())
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
	bs, err := a.conn.Encode(param)
	if err != nil {
		return err
	}

	if err := a.conn.Write(bs); err != nil {
		return err
	}

	for {
		body, err := a.conn.Decode(a.conn.Reader())
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
