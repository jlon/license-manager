package service

import (
	"context"
	"errors"
	"testing"

	"license-manager/internal/models"
	"license-manager/internal/repository"
	"license-manager/pkg/i18n"
)

type deleteLicenseRepositoryStub struct {
	license *models.License
	getErr  error
	deleted bool
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
	return nil, nil
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
