package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"license-manager/internal/models"
	"license-manager/internal/repository"
	"license-manager/pkg/i18n"
)

type deleteLicenseRepositoryStub struct {
	license      *models.License
	licenseByKey *models.License
	getErr       error
	deleted      bool
	updated      bool
}

func (r *deleteLicenseRepositoryStub) GetLicenseList(context.Context, *models.LicenseListRequest) (*models.LicenseListResponse, error) {
	return nil, nil
}
func (r *deleteLicenseRepositoryStub) GetLicenseByID(context.Context, string) (*models.License, error) {
	return r.license, r.getErr
}
func (r *deleteLicenseRepositoryStub) CreateLicense(context.Context, *models.License) error {
	return nil
}
func (r *deleteLicenseRepositoryStub) UpdateLicense(context.Context, *models.License) error {
	r.updated = true
	return nil
}
func (r *deleteLicenseRepositoryStub) DeleteLicensePermanently(context.Context, *models.License) error {
	r.deleted = true
	return nil
}
func (r *deleteLicenseRepositoryStub) CheckAuthorizationCodeExists(context.Context, string) (bool, error) {
	return false, nil
}
func (r *deleteLicenseRepositoryStub) GetAuthorizationCodeByID(context.Context, string) (*models.AuthorizationCode, error) {
	return nil, nil
}
func (r *deleteLicenseRepositoryStub) GetAuthorizationCodeByCode(context.Context, string) (*models.AuthorizationCode, error) {
	return nil, nil
}
func (r *deleteLicenseRepositoryStub) GetLicenseByKey(context.Context, string) (*models.License, error) {
	return r.licenseByKey, r.getErr
}

func TestStellarTrialAuthorizationCodeIsStableAndScoped(t *testing.T) {
	secret := "0123456789abcdef0123456789abcdef"
	fingerprint := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	first := stellarTrialAuthorizationCode(secret, fingerprint)
	second := stellarTrialAuthorizationCode(secret, fingerprint)
	if first != second {
		t.Fatalf("trial code must be deterministic")
	}
	if first == stellarTrialAuthorizationCode(secret, strings.Repeat("f", 64)) {
		t.Fatalf("trial code must change when fingerprint changes")
	}
	if !isSHA256Hex(fingerprint) || isSHA256Hex("not-a-fingerprint") {
		t.Fatalf("SHA-256 fingerprint validation is incorrect")
	}
}

func TestHeartbeatRejectsMismatchedFingerprint(t *testing.T) {
	repo := &deleteLicenseRepositoryStub{
		licenseByKey: &models.License{
			LicenseKey:          "license-1",
			HardwareFingerprint: strings.Repeat("a", 64),
			Status:              "active",
		},
	}
	service := &licenseService{licenseRepo: repo}
	_, err := service.Heartbeat(context.Background(), &models.HeartbeatRequest{
		LicenseKey:          "license-1",
		HardwareFingerprint: strings.Repeat("b", 64),
	}, "127.0.0.1")

	var i18nErr *i18n.I18nError
	if !errors.As(err, &i18nErr) || i18nErr.Code != "300006" {
		t.Fatalf("got error %v, want license-not-found code", err)
	}
	if repo.updated {
		t.Fatal("heartbeat must not update a license after fingerprint mismatch")
	}
}
func (r *deleteLicenseRepositoryStub) GetActiveLicenseCount(context.Context, string) (int64, error) {
	return 0, nil
}
func (r *deleteLicenseRepositoryStub) CheckAuthorizationCodeHasLicenses(context.Context, string) (bool, error) {
	return false, nil
}

func TestDeleteLicenseOnlyAllowsRevokedStatus(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		getErr      error
		wantCode    string
		wantDeleted bool
	}{
		{name: "revoked license", status: "revoked", wantDeleted: true},
		{name: "active license", status: "active", wantCode: "300012"},
		{name: "inactive license", status: "inactive", wantCode: "300012"},
		{name: "missing license", getErr: repository.ErrLicenseNotFound, wantCode: "300006"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &deleteLicenseRepositoryStub{getErr: tt.getErr}
			if tt.getErr == nil {
				repo.license = &models.License{ID: "license-1", Status: tt.status}
			}
			service := &licenseService{licenseRepo: repo}
			err := service.DeleteLicense(context.Background(), "license-1")

			if tt.wantCode == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				var i18nErr *i18n.I18nError
				if !errors.As(err, &i18nErr) || i18nErr.Code != tt.wantCode {
					t.Fatalf("got error %v, want code %s", err, tt.wantCode)
				}
			}
			if repo.deleted != tt.wantDeleted {
				t.Fatalf("deleted = %v, want %v", repo.deleted, tt.wantDeleted)
			}
		})
	}
}
