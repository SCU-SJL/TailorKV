package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"net"
)

func doAuth(conn net.Conn, login *AESLogin) bool {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return false
	}
	encrypted := string(buf[:n])
	decrypted, err := AesDecrypt(encrypted, login.AESKey)
	return err == nil && login.AuthPassword == decrypted
}

func auth(conn net.Conn, login *AESLogin) error {
	if !login.AuthRequired {
		_, err := conn.Write([]byte{0})
		return err
	}

	_, err := conn.Write([]byte{1})
	if err != nil {
		return err
	}

	login.AuthPassed = doAuth(conn, login)
	if !login.AuthPassed {
		_, _ = conn.Write([]byte{1})
		return errors.New("password is wrong")
	}
	_, err = conn.Write([]byte{0})
	return err
}

func AesDecrypt(encrypted string, key string) (string, error) {
	encryptedByte, _ := base64.StdEncoding.DecodeString(encrypted)
	k := []byte(key)

	block, _ := aes.NewCipher(k)
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	orig := make([]byte, len(encryptedByte))
	blockMode.CryptBlocks(orig, encryptedByte)
	orig, err := PKCS7UnPadding(orig)

	return string(orig), err
}

func PKCS7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	padLen := int(origData[length-1])
	if length < padLen {
		return nil, errors.New("padding length is wrong")
	}
	return origData[:(length - padLen)], nil
}
