package util

import "golang.org/x/crypto/bcrypt"

// PasswordHash implements root.Hash
type PasswordHash struct{}

func (c *PasswordHash) GenerateHashFromString(s string) (string, error) {
	saltedBytes := []byte(s)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hash := string(hashedBytes[:])

	return hash, nil
}

func (c *PasswordHash) CheckHashEquality(hashedPassword string, password string) error {
	hashedPasswordBytes := []byte(hashedPassword)
	passwordBytes := []byte(password)

	return bcrypt.CompareHashAndPassword(hashedPasswordBytes, passwordBytes)
}
