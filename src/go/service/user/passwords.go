package user

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

const defaultMemory uint32 = 64 * 1024
const defaultTimeCost uint32 = 3
const defaultParallelization uint8 = 1

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}
	saltBase64 := base64.RawStdEncoding.Strict().EncodeToString(salt)

	digest := argon2.IDKey([]byte(password), salt, defaultTimeCost, defaultMemory, defaultParallelization, 32)
	digestBase64 := base64.RawStdEncoding.Strict().EncodeToString(digest)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, defaultMemory, defaultTimeCost, defaultParallelization, saltBase64, digestBase64), nil
}

func VerifyPassword(password string, storedHash string) (bool, error) {
	phcSegments := strings.Split(storedHash, "$")
	if len(phcSegments) != 6 {
		return false, errors.New("invalid stored hash")
	}

	if phcSegments[1] != "argon2id" {
		return false, fmt.Errorf("invalid algorithm: %s", phcSegments[1])
	}

	var version int
	_, err := fmt.Sscanf(phcSegments[2], "v=%d", &version)
	if err != nil {
		return false, fmt.Errorf("unable to parse version: %s caused by: %w", phcSegments[2], err)
	}

	if version != argon2.Version {
		return false, fmt.Errorf("incompatible version: %d", version)
	}

	var memory uint32
	var timeCost uint32
	var parallelization uint8
	_, err = fmt.Sscanf(phcSegments[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &parallelization)
	if err != nil {
		return false, fmt.Errorf("unable to parse parameters: %s caused by: %w", phcSegments[3], err)
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(phcSegments[4])
	if err != nil {
		return false, fmt.Errorf("unable to decode salt caused by: %w", err)
	}

	digest, err := base64.RawStdEncoding.Strict().DecodeString(phcSegments[5])
	if err != nil {
		return false, fmt.Errorf("unable to decode digest caused by: %w", err)
	}
	keyLength := uint32(len(digest))

	inputDigest := argon2.IDKey([]byte(password), salt, timeCost, memory, parallelization, keyLength)

	if subtle.ConstantTimeCompare(digest, inputDigest) == 1 {
		return true, nil
	}
	return false, nil
}
