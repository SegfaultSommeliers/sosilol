package utils

import "crypto/rand"

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomID(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}

	b := make([]byte, n)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	out := make([]byte, n)

	for i := 0; i < n; i++ {
		out[i] = charset[b[i]%byte(len(charset))]
	}

	return string(out), nil
}
