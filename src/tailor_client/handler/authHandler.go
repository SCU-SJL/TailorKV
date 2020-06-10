package handler

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/howeyc/gopass"
	"log"
	"net"
)

func auth(conn net.Conn) error {
	ok, err := authorized(conn)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		password := readAuth("password")
		aesKey := readAuth("AES key")
		password = AesEncrypt(password, aesKey)

		_, err = conn.Write([]byte(password))
		if err != nil {
			log.Fatal(err)
		}

		resp := make([]byte, 1)
		_, err = conn.Read(resp)
		if err != nil {
			log.Fatal(err)
		}

		if resp[0] != 0 {
			return errors.New("wrong password")
		}
	}
	return nil
}

func readAuth(param string) string {
	fmt.Printf("Enter the %s: ", param)

	input, err := gopass.GetPasswdMasked()

	for err != nil {
		fmt.Println("invalid input")
		fmt.Printf("Enter %s again: ", param)
		input, err = gopass.GetPasswdMasked()
	}
	return string(input)
}

func authorized(conn net.Conn) (bool, error) {
	buf := make([]byte, 1)
	_, err := conn.Read(buf)
	if err != nil {
		return false, err
	}
	if buf[0] == 0 {
		return true, nil
	}
	return false, nil
}

func AesEncrypt(orig string, key string) string {
	origData := []byte(orig)
	k := []byte(key)

	block, err := aes.NewCipher(k)
	if err != nil {
		log.Fatal(err)
	}

	blockSize := block.BlockSize()
	origData = PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)

	return base64.StdEncoding.EncodeToString(encrypted)
}

func PKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}
