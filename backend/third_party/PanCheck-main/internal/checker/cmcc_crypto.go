package checker

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// chinaMobileCloudDecrypt 中国移动云解密函数
func chinaMobileCloudDecrypt(encryptedText string) (string, error) {
	// 固定密钥
	key := []byte("PVGDwmcvfs1uV3d1")

	// Base64解码
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %v", err)
	}

	// 检查数据长度，至少需要16字节（一个AES块）
	if len(encryptedData) < 16 {
		return "", fmt.Errorf("加密数据长度不足")
	}

	// 前16字节作为IV，后面的作为加密数据
	iv := encryptedData[:16]
	ciphertext := encryptedData[16:]

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %v", err)
	}

	// 检查密文长度必须是块大小的倍数
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("密文长度不是块大小的倍数")
	}

	// 创建CBC模式的解密器
	mode := cipher.NewCBCDecrypter(block, iv)

	// 解密
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除PKCS7填充
	plaintext, err = removePKCS7Padding(plaintext)
	if err != nil {
		return "", fmt.Errorf("去除填充失败: %v", err)
	}

	return string(plaintext), nil
}

// removePKCS7Padding 去除PKCS7填充
func removePKCS7Padding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("数据为空")
	}

	paddingLength := int(data[len(data)-1])
	if paddingLength == 0 || paddingLength > len(data) {
		return nil, fmt.Errorf("无效的填充长度")
	}

	// 验证填充
	for i := len(data) - paddingLength; i < len(data); i++ {
		if data[i] != byte(paddingLength) {
			return nil, fmt.Errorf("填充验证失败")
		}
	}

	return data[:len(data)-paddingLength], nil
}

// addPKCS7Padding 添加PKCS7填充
func addPKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// chinaMobileCloudEncrypt 中国移动云加密函数
func chinaMobileCloudEncrypt(data interface{}) (string, error) {
	// 固定密钥
	key := []byte("PVGDwmcvfs1uV3d1")

	// 生成16字节的随机IV
	iv := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("生成IV失败: %v", err)
	}

	// 准备要加密的数据
	var plaintext []byte
	var err error

	switch v := data.(type) {
	case string:
		plaintext = []byte(v)
	case []byte:
		plaintext = v
	default:
		// 对于其他类型，尝试转换为JSON
		if reflect.TypeOf(v).Kind() == reflect.Struct || reflect.TypeOf(v).Kind() == reflect.Map {
			plaintext, err = json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("JSON序列化失败: %v", err)
			}
		} else {
			return "", fmt.Errorf("不支持的数据类型: %T", v)
		}
	}

	// 添加PKCS7填充
	plaintext = addPKCS7Padding(plaintext, aes.BlockSize)

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %v", err)
	}

	// 创建CBC模式的加密器
	mode := cipher.NewCBCEncrypter(block, iv)

	// 加密
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)

	// 将IV和密文连接
	result := append(iv, ciphertext...)

	// Base64编码
	return base64.StdEncoding.EncodeToString(result), nil
}
