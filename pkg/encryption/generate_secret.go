package encryption

import (
    "crypto/rand"
    "math/big"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()_+-=[]{}|;:,.<>?")

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
