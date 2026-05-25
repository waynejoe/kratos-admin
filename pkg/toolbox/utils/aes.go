package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// AesEcbEncrypt 采用ECB模式加密
// @description: ECB加密后，两相同明文的密文也完全相同！模式是不安全的!
func AesEcbEncrypt(text, key string) (string, error) {
	plaintext := []byte(text)
	plaintkey := []byte(key)
	block, err := aes.NewCipher(plaintkey)
	if err != nil {
		return "", err
	}

	// 填充明文
	blockSize := block.BlockSize()
	paddedPlaintext := pkcs7Pad(plaintext, blockSize)

	// ECB 加密：每个块独立加密
	ciphertext := make([]byte, len(paddedPlaintext))
	for i := 0; i < len(paddedPlaintext); i += blockSize {
		block.Encrypt(ciphertext[i:], paddedPlaintext[i:i+blockSize])
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AesEcbDecrypt 采用ECB模式解密
func AesEcbDecrypt(text, key string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}
	cipherkey := []byte(key)
	block, err := aes.NewCipher(cipherkey)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	if len(ciphertext)%blockSize != 0 {
		return "", fmt.Errorf("密文长度不是块大小的倍数")
	}

	// ECB 解密：每个块独立解密
	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += blockSize {
		block.Decrypt(plaintext[i:], ciphertext[i:i+blockSize])
	}

	plaintext = pkcs7Unpad(plaintext) // 去除填充
	return string(plaintext), nil
}

// AesDecrypt 解密
// description: cbc/gcm更安全，但密文生成每次都不相同
func AesDecrypt(text, key string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}
	cipherKey := []byte(key)
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// 去除填充
	ciphertext = pkcs7Unpad(ciphertext)

	return string(ciphertext), nil
}

// AES 加密
func AesEncrypt(text, key string) (string, error) {
	plaintext := []byte(text)
	plaintkey := []byte(key)
	block, err := aes.NewCipher(plaintkey)
	if err != nil {
		return "", err
	}

	// 填充明文到块大小的倍数
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)

	// IV 长度应该等于块大小
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// PKCS7 填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// PKCS7 去除填充
func pkcs7Unpad(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}
