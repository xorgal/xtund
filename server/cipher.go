// File: server/cipher.go
package server

import (
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

// Todo:
// 1. Synchronize nonce between server/client
// 2. Implement Encrypt/Descrypt functions

const (
	KeySize       = 32
	NonceSize     = 12
	Overhead      = 16
	PBKDFSaltSize = 16
	PBKDF2I       = 4096
)

func DeriveEncryptionKey(source string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(source))
	salt := hasher.Sum(nil)[:PBKDFSaltSize]

	b := []byte(source)

	return pbkdf2.Key(b, salt, PBKDF2I, KeySize, sha256.New)
}

// Encryption
func Encrypt(data []byte, key []byte) {

}

// Decryption
func Decrypt(data []byte, key []byte) {

}
