package repository

import (
	"strings"
	"testing"
	"time"

	"license-manager/internal/models"
)

func TestAuthorizationDeviceFilterKeepsPredicatesOnOneDevice(t *testing.T) {
	online := true
	threshold := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	filter, args := authorizationDeviceFilter(&models.AuthorizationCodeListRequest{
		HardwareFingerprint: "machine-123",
		LicenseStatus:       "active",
		DeviceOnline:        &online,
	}, threshold)

	for _, condition := range []string{
		"EXISTS (SELECT 1 FROM licenses WHERE",
		"licenses.authorization_code_id = ac.id",
		"licenses.hardware_fingerprint LIKE ? ESCAPE '!'",
		"licenses.status = ?",
		"licenses.last_heartbeat >= ?",
	} {
		if !strings.Contains(filter, condition) {
			t.Fatalf("filter %q is missing %q", filter, condition)
		}
	}
	if len(args) != 3 || args[0] != "machine-123%" || args[1] != "active" || args[2] != threshold {
		t.Fatalf("unexpected filter arguments: %#v", args)
	}
}

func TestAuthorizationDeviceFilterEscapesFingerprintWildcards(t *testing.T) {
	filter, args := authorizationDeviceFilter(&models.AuthorizationCodeListRequest{
		HardwareFingerprint: "machine_%!",
	}, time.Time{})

	if !strings.Contains(filter, "licenses.hardware_fingerprint LIKE ? ESCAPE '!'") {
		t.Fatalf("fingerprint filter is incorrect: %q", filter)
	}
	if len(args) != 1 || args[0] != "machine!_!%!!%" {
		t.Fatalf("unexpected fingerprint filter arguments: %#v", args)
	}
}

func TestAuthorizationDeviceFilterMatchesOfflineDevices(t *testing.T) {
	offline := false
	filter, args := authorizationDeviceFilter(&models.AuthorizationCodeListRequest{
		DeviceOnline: &offline,
	}, time.Time{})

	if !strings.Contains(filter, "(licenses.last_heartbeat < ? OR licenses.last_heartbeat IS NULL)") {
		t.Fatalf("offline filter is incorrect: %q", filter)
	}
	if len(args) != 1 {
		t.Fatalf("unexpected offline filter arguments: %#v", args)
	}
}
