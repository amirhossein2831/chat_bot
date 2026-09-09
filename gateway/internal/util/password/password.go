package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	time    = 1
	memory  = 64 * 1024
	threads = 4
	keyLen  = 32
	saltLen = 16

	prefix = "argon2id"
)

var ErrMalformedHash = errors.New("malformed password hash")

func Hash(plain string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(plain), salt, time, memory, threads, keyLen)

	return fmt.Sprintf("%s$%d$%d$%d$%s$%s",
		prefix, memory, time, threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

// Verify reports whether plain produced encoded. It compares in constant time
// so a timing signal cannot reveal how much of the hash matched.
func Verify(plain, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != prefix {
		return false, ErrMalformedHash
	}

	var mem uint32
	var iter, par int
	if _, err := fmt.Sscanf(strings.Join(parts[1:4], " "), "%d %d %d", &mem, &iter, &par); err != nil {
		return false, ErrMalformedHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrMalformedHash
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, ErrMalformedHash
	}

	got := argon2.IDKey([]byte(plain), salt, uint32(iter), mem, uint8(par), uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
