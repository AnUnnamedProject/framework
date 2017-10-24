package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"reflect"
	"testing"
)

// func BenchmarkEncrypt(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		Encrypt([]byte("92kdfe2mxd232"), []byte("TEST"))
// 	}
// }

// func BenchmarkEncryptBase64(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		e, _ := Encrypt([]byte("92kdfe2mxd232"), []byte("TEST"))
// 		EncodeBase64(e)
// 	}
// }

// func BenchmarkSignature(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		createSignature("TEST", []byte("HOLA"))
// 	}
// }

func BenchmarkEncryptXOR(b *testing.B) {
	k := []byte("TEST")
	s := []byte("PRIVATE STRING")
	for i := 0; i < b.N; i++ {
		EncryptXOR(k, s)
	}
}

func BenchmarkDecryptXOR(b *testing.B) {
	k := []byte("TEST")
	s := []byte("PRIVATE STRING")
	e := EncryptXOR(k, s)

	for i := 0; i < b.N; i++ {
		DecryptXOR(k, e)
	}
}

func TestDecryptXOR(t *testing.T) {
	k := []byte("TEST")
	s := []byte("2409hj2302jwf020jt0923h0t32h290323h09t2390")

	encrypted := EncryptXOR(k, s)
	decrypted := DecryptXOR(k, encrypted)

	if !reflect.DeepEqual(decrypted, s) {
		t.Errorf("DecryptXOR is expected to be %v, got %v", s, decrypted)
	}
}

func createSignature(secret string, parts ...[]byte) []byte {
	h := hmac.New(sha1.New, []byte(secret))

	for _, x := range parts {
		h.Write(x)
	}

	hexDigest := make([]byte, 64)
	hex.Encode(hexDigest, h.Sum(nil))

	return hexDigest[:bytes.Index(hexDigest, []byte("\000"))]
}

func charCodeAt(s []byte, n int) int {
	if n > len(s) {
		return 0
	}

	return int(rune(s[n]))
}

// func BenchmarkDecrypt(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		Decrypt([]byte("92kdfe2mxd232"), []byte("TEST"))
// 	}
// }

func BenchmarkRandomStringAlnum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandomString("alnum", 8)
	}
}
