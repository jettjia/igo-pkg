package tls

import "testing"

// go test -v -run Test_GenerateKeyTls ./
func Test_GenerateKeyTls(t *testing.T) {
	key, err := GenerateKeyTls()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(key)
}

// go test -v -run Test_GenerateCertTls ./
func Test_GenerateCertTls(t *testing.T) {
	key, err := GenerateKeyTls()
	if err != nil {
		t.Error(err)
		return
	}
	cert, err := GenerateCertTls(key, "xxx.tech", []string{"xxx Tech"})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(cert)
}
