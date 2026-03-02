package license

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/hyperboloide/lk"
)

// LicenseDetails holds the details of a license.
type LicenseDetails struct {
	Company       string    `json:"company"`        // company
	End           time.Time `json:"end"`            // expiration date
	LicenseNumber string    `json:"license_number"` // license number
	Features      []string  `json:"features"`       // purchased features
}

// GenerateLicense generates a license from a private key.
func GenerateLicense(privateKeyBase32 string, details LicenseDetails) (string, error) {
	docBytes, err := json.Marshal(details)
	if err != nil {
		return "", err
	}

	privateKey, err := lk.PrivateKeyFromB32String(privateKeyBase32)
	if err != nil {
		return "", err
	}

	license, err := lk.NewLicense(privateKey, docBytes)
	if err != nil {
		return "", err
	}

	licenseB32, err := license.ToB32String()
	if err != nil {
		return "", err
	}

	return licenseB32, nil
}

// VerifyLicense validates a license with a public key and returns the license details.
func VerifyLicense(licenseB32 string, publicKeyBase32 string) (bool, LicenseDetails, error) {
	publicKey, err := lk.PublicKeyFromB32String(publicKeyBase32)
	if err != nil {
		return false, LicenseDetails{}, err
	}

	license, err := lk.LicenseFromB32String(licenseB32)
	if err != nil {
		return false, LicenseDetails{}, err
	}

	if ok, err := license.Verify(publicKey); err != nil {
		return false, LicenseDetails{}, err
	} else if !ok {
		return false, LicenseDetails{}, nil
	}

	var details LicenseDetails
	if err := json.Unmarshal(license.Data, &details); err != nil {
		return false, LicenseDetails{}, err
	}

	if details.End.Before(time.Now()) {
		return false, details, nil
	}

	return true, details, nil
}

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateCode generates a random string of characters in the specified format.
// 改用 crypto/rand，避免伪随机数暴露安全风险。
func GenerateCode(length int, numParts int) (string, error) {
	if length <= 0 || numParts <= 0 {
		return "", errors.New("length and numParts must be positive")
	}

	var b strings.Builder
	for i := 0; i < numParts; i++ {
		part, err := generatePart(length)
		if err != nil {
			return "", err
		}
		if i > 0 {
			b.WriteByte('-')
		}
		b.WriteString(part)
	}
	return b.String(), nil
}

// generatePart generates a random string of individual parts using crypto/rand.
func generatePart(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = letterBytes[int(b[i])%len(letterBytes)]
	}
	return string(b), nil
}
