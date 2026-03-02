package license

import (
	"os"
	"testing"
	"time"
)

// go test -v -run Test_license ./
func Test_license(t *testing.T) {
	privateKeyBase32 := os.Getenv("LICENSE_PRIVATE_KEY_B32")
	publicKeyBase32 := os.Getenv("LICENSE_PUBLIC_KEY_B32")
	if privateKeyBase32 == "" || publicKeyBase32 == "" {
		t.Skip("missing LICENSE_PRIVATE_KEY_B32 or LICENSE_PUBLIC_KEY_B32; skip to avoid exposing test keys")
	}

	num, err := GenerateCode(5, 4)
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}
	details := LicenseDetails{
		Company:       "上海东方明珠",
		End:           time.Now().Add(time.Hour * 24 * 365),
		LicenseNumber: num,
		Features: []string{
			"xtext",
		},
	}

	licenseB32, err := GenerateLicense(privateKeyBase32, details)
	if err != nil {
		t.Fatalf("failed to generate license: %v", err)
	}
	t.Logf("generated license length=%d", len(licenseB32))

	isValid, _, err := VerifyLicense(licenseB32, publicKeyBase32)
	if err != nil {
		t.Fatalf("failed to verify license: %v", err)
	}
	if !isValid {
		t.Fatalf("license should be valid")
	}
}

// go test -v -run Test_VerifyLicense ./
func Test_VerifyLicense(t *testing.T) {
	licenseB32 := os.Getenv("LICENSE_B32")
	publicKeyBase32 := os.Getenv("LICENSE_PUBLIC_KEY_B32")
	if licenseB32 == "" || publicKeyBase32 == "" {
		t.Skip("missing LICENSE_B32 or LICENSE_PUBLIC_KEY_B32; skip to avoid exposing test data")
	}

	isValid, _, err := VerifyLicense(licenseB32, publicKeyBase32)
	if err != nil {
		t.Fatalf("failed to verify license: %v", err)
	}

	if !isValid {
		t.Fatalf("license should be valid")
	}
}
