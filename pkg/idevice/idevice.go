package idevice

import "errors"

func ConnectLockdownWithSession(entry *DeviceEntry) (*LockdownConn, error) {
	conn, err := NewUSBConn()
	if err != nil {
		return nil, err
	}
	// defer conn.Close()

	cert, err := conn.GetCertificate(entry.Properties.SerialNumber)
	if err != nil {
		return nil, err
	}

	lockdown, err := conn.ConnectLockdown(entry.DeviceID)
	if err != nil {
		return nil, err
	}

	_, err = lockdown.StartSession(cert)
	if err != nil {
		return nil, err
	}

	return lockdown, nil
}

func StartService(entry *DeviceEntry, name string) (*StartServiceResponse, error) {
	lockdown, err := ConnectLockdownWithSession(entry)
	if err != nil {
		return nil, err
	}
	defer lockdown.Close()

	return lockdown.StartService(name)
}

func GetCertificate(udid string) (*Certificate, error) {
	conn, err := NewUSBConn()
	if err != nil {
		return nil, err
	}

	return conn.GetCertificate(udid)
}

func ConnectToService(entry *DeviceEntry, name string) (IConn, error) {
	resp, err := StartService(entry, name)
	if err != nil {
		return nil, err
	}

	cert, err := GetCertificate(entry.Properties.SerialNumber)
	if err != nil {
		return nil, err
	}

	conn, err := NewUSBConn()
	if err != nil {
		return nil, err
	}

	if err := conn.ConnectWithStartServiceResponse(entry.DeviceID, resp, cert); err != nil {
		return nil, err
	}

	return conn.Conn, nil
}

func GetDevice(udid ...string) (device *DeviceEntry, err error) {
	conn, err := NewUSBConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	list, err := conn.ListDevices()
	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		device = &list[0]
		if len(udid) > 0 {
			for _, entry := range list {
				if udid[0] == entry.Properties.SerialNumber {
					device = &entry
					break
				}
			}
		}

		return
	}

	return nil, errors.New("没有连接任何iOS设备")
}
