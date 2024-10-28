package encryption

import (
    "crypto/rand"
    "math/big"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()_+-=[]{}|;:,.<>?")

// Generate random string this can provide 88^n possible
func RandStringRunes(n int) (string, error) {
	b := make([]rune, n)
	for i := range b {
        index, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterRunes))))
        if err != nil {
            return "", err
        }
        b[i] = letterRunes[index.Int64()]
	}
	return string(b), nil
}
