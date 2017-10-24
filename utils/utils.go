package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"io"
	"math"
	rnd "math/rand"
	"strconv"
	"strings"
	"time"
)

// StrPad returns the input string padded on the left, right or both sides using padType to the specified padding length padLength.
//
// Example:
// input := "Codes";
// StrPad(input, 10, " ", "RIGHT")        // produces "Codes     "
// StrPad(input, 10, "-=", "LEFT")        // produces "=-=-=Codes"
// StrPad(input, 10, "_", "BOTH")         // produces "__Codes___"
// StrPad(input, 6, "___", "RIGHT")       // produces "Codes_"
// StrPad(input, 3, "*", "RIGHT")         // produces "Codes"
func StrPad(input string, padLength int, padString string, padType string) string {
	var output string

	inputLength := len(input)
	padStringLength := len(padString)

	if inputLength >= padLength {
		return input
	}

	repeat := math.Ceil(float64(1) + (float64(padLength-padStringLength))/float64(padStringLength))

	switch padType {
	case "RIGHT":
		output = input + strings.Repeat(padString, int(repeat))
		output = output[:padLength]
	case "LEFT":
		output = strings.Repeat(padString, int(repeat)) + input
		output = output[len(output)-padLength:]
	case "BOTH":
		length := (float64(padLength - inputLength)) / float64(2)
		repeat = math.Ceil(length / float64(padStringLength))
		output = strings.Repeat(padString, int(repeat))[:int(math.Floor(float64(length)))] + input + strings.Repeat(padString, int(repeat))[:int(math.Ceil(float64(length)))]
	}

	return output
}

// EncodeBase64 returns an encoded base64 string from []byte.
func EncodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// DecodeBase64 returns a decode base64 []byte from string.
func DecodeBase64(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

// Encrypt return an AES encrypted []byte.
func Encrypt(block cipher.Block, key, text []byte) ([]byte, error) {
	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return ciphertext, nil
}

// Decrypt returns an AES decrypted string.
func Decrypt(block cipher.Block, key, value []byte) []byte {
	if len(value) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := value[:aes.BlockSize]
	value = value[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(value, value)
	return value
}

// EncryptXOR return a XORed encrypted string.
func EncryptXOR(key, value []byte) string {
	value = append([]byte("0"), value...)
	valueLen := len(value)
	keyLen := len(key)
	random := 65
	count := valueLen + (random % keyLen)
	values := make([]int, count)
	values[0] = random

	for i := count - 1; i > 0; i-- {
		values[i] = int(rune(value[i%valueLen])) ^ (int(rune(key[i%keyLen])) ^ random)
	}

	encoded := &bytes.Buffer{}
	encoded.WriteString(strconv.Itoa(valueLen + keyLen))
	encoded.WriteString("=")

	for _, v := range values {
		encoded.WriteString(string(v))
	}

	return EncodeBase64(encoded.Bytes())
}

// DecryptXOR return a decrypted XORed string.
func DecryptXOR(key []byte, value string) []byte {
	decoded := DecodeBase64(value)

	index := bytes.Index(decoded, []byte("="))
	if index == -1 {
		return nil
	}

	_counter, err := strconv.ParseInt(string(decoded[0:index]), 10, 64)
	if err != nil {
		return nil
	}

	counter := int(_counter)
	decoded = decoded[index+1:]

	l := len(decoded)
	random := int(rune(decoded[0]))
	keyLen := len(key)
	valueLen := l - (random % keyLen)

	values := make([]int, valueLen-1)

	for i := valueLen - 1; i > 0; i-- {
		values[i-1] = int(rune(decoded[i])) ^ (random ^ int(rune(key[i%keyLen])))
	}

	if counter != len(values)+1+keyLen {
		return nil
	}

	decrypted := &bytes.Buffer{}
	for _, v := range values {
		decrypted.WriteString(string(v))
	}

	return decrypted.Bytes()
}

// GobEncode return an encoded golang gob []byte
func GobEncode(data map[string]interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GobDecode return a pointer of decoded gob []byte
func GobDecode(data []byte, values *map[string]interface{}) error {
	buf := bytes.NewBuffer(data)
	return gob.NewDecoder(buf).Decode(&values)
}

func RandomString(t string, l int) string {
	b := make([]byte, l)
	r := rnd.New(rnd.NewSource(time.Now().Unix()))

	pool := ""

	switch t {
	case "alnum":
		pool = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case "alpha":
		pool = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case "hexdec":
		pool = "0123456789abcdef"
	case "numeric":
		pool = "0123456789"
	case "nozero":
		pool = "123456789"
	case "distinct":
		pool = "2345679ACDEFHJKLMNPRSTUVWXYZ"
	}

	// Largest pool key
	max := len(pool) - 1

	for i := 0; i < l; i++ {
		b[i] = pool[r.Intn(max)]
	}

	return string(b)
}
