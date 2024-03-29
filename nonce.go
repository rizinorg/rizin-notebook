package main

import (
	"crypto/rand"
)

const ELEMENT_NONCE_SIZE = 8
const PAGE_NONCE_SIZE = 32

func Nonce(n int) string {
	const encoding = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const modulo = byte(len(encoding))
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	for i, b := range bytes {
		bytes[i] = encoding[b%modulo]
	}
	return string(bytes)
}

func IsValidNonce(unique string) bool {
	return true
}
