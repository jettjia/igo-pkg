package xcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// AesEncrypt encrypt
func AesEncrypt(key, data []byte) ([]byte, error) {
	//create an encrypted instance
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//determine the size of the encrypted block
	blockSize := block.BlockSize()
	//padding
	encryptBytes := pkcs7Padding(data, blockSize)
	//initialize the encrypted data receive slices
	crypted := make([]byte, len(encryptBytes))
	//use cbc encryption mode
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	//perform encryption
	blockMode.CryptBlocks(crypted, encryptBytes)
	return crypted, nil
}

// AesDecrypt decrypt
func AesDecrypt(key, data []byte) ([]byte, error) {
	//create an instance
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//get the size of the block
	blockSize := block.BlockSize()
	//use cbc
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	//initial decryption of data receive slices
	crypted := make([]byte, len(data))
	//perform decryption
	blockMode.CryptBlocks(crypted, data)
	//remove the filling
	crypted, err = pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}
	return crypted, nil
}

// pkcs7Padding padding
func pkcs7Padding(data []byte, blockSize int) []byte {
	//judge the missing digits in length the minimum is 1 and the maximum block size is maximum
	padding := blockSize - len(data)%blockSize
	//make up the digits copy the slice byte byte padding to the padding
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7UnPadding reverse operation of the fill
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7UnPadding err")
	}
	//get the number of fills
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}
