package xcrypto

import (
	"encoding/base64"
	"testing"
)

// go test -v -run=Test_AesEncrypt .
func Test_AesEncrypt(t *testing.T) {
	key := []byte("1234567891234567")
	data := []byte("123Abc")

	encrypted, err := AesEncrypt(key, data)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(base64.StdEncoding.EncodeToString(encrypted)) // b1sa3E/Q700QRu8lXcqFDA==
}

// go test -v -run=Test_AesDecrypt .
func Test_AesDecrypt(t *testing.T) {
	key := []byte("1234567891234567")
	data := "b1sa3E/Q700QRu8lXcqFDA=="

	deData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		t.Fatal(err)
	}

	decrypt, err := AesDecrypt(key, deData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(decrypt)) //123Abc
}
