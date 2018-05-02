package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
)

// Encrypt 加密函数
// 参数 plainText 待加密字符串 key : 加密使用的key
// 返回 加密后的字符串
func Encrypt(plainText string, key string) (cipherText string, err error) {
	var block cipher.Block
	keyBytes := hashBytes(key)
	plainTextBytes := []byte(plainText)
	block, err = aes.NewCipher(keyBytes)
	if err != nil {
		return
	}

	cipherTextBytes := make([]byte, aes.BlockSize+len(plainTextBytes))
	iv := cipherTextBytes[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherTextBytes[aes.BlockSize:], plainTextBytes)
	cipherText = "crypt-" + hex.EncodeToString(cipherTextBytes)
	return
}

// Decrypt 解密函数
// 参数 cipherText 待解密的字符串， key: 解密使用的key
// 需要使用与加密相同的key才能正确的解密，这也是算法保密性的关键
// 返回 解密后的信息
func Decrypt(cipherText string, key string) (plainText string, err error) {
	if len(cipherText) == 0 || len(cipherText) < 6 || cipherText[:6] != "crypt-" {
		err = errors.New("Illegal ciphertext")
		return
	}
	cipherText = string(cipherText[6:])
	var block cipher.Block
	keyBytes := hashBytes(key)
	cipherTextBytes, _ := hex.DecodeString(cipherText)
	block, err = aes.NewCipher(keyBytes)
	if err != nil {
		return
	}

	if len(cipherTextBytes) < aes.BlockSize {
		err = errors.New("Ciphertext too short")
		return
	}

	iv := cipherTextBytes[:aes.BlockSize]
	cipherTextBytes = cipherTextBytes[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)

	plainTextBytes := make([]byte, len(cipherTextBytes))
	stream.XORKeyStream(plainTextBytes, cipherTextBytes)
	plainText = string(plainTextBytes)
	return
}

func hashBytes(key string) (hash []byte) {
	h := sha1.New()
	io.WriteString(h, key)
	hashStr := hex.EncodeToString(h.Sum(nil))
	hash = []byte(hashStr)[:32]
	return
}
