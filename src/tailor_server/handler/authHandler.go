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
	decrypted := AesDecrypt(encrypted, login.AESKey)
	return login.AuthPassword == decrypted
}

func auth(conn net.Conn, login *AESLogin) error {
	if !login.AuthRequired {
		_, err := conn.Write([]byte{0})
		return err
	}

	_, err := conn.Write([]byte{1})
	if err != nil {
		return nil
	}

	login.AuthPassed = doAuth(conn, login)
	if !login.AuthPassed {
		return errors.New("password is wrong")
	}
	return nil
}

func AesDecrypt(encrypted string, key string) string {
	encryptedByte, _ := base64.StdEncoding.DecodeString(encrypted)
	k := []byte(key)

	block, _ := aes.NewCipher(k)
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	orig := make([]byte, len(encryptedByte))
	blockMode.CryptBlocks(orig, encryptedByte)
	orig = PKCS7UnPadding(orig)

	return string(orig)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	padLen := int(origData[length-1])
	return origData[:(length - padLen)]
}
