package main

import "crypto/rand"

var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ") 

func randomString(length int) string {
	if length <= 0 {
		return ""
	}

	outputBytes := make([]rune, length)
	randBytes := make([]byte, length)
	rand.Read(randBytes)
	for i := range outputBytes {
		j := int(randBytes[i]) % len(letters)
		outputBytes[i] = letters[j]
	}
	return string(outputBytes)
}