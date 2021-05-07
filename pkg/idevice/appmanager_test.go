package idevice

import "testing"

func TestAppManagerService_GetApplications(t *testing.T) {
	device, err := GetDevice()
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewAppManagerService(device)
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()

	apps, err := service.GetApplications()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(apps)
}
