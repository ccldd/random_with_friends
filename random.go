package main

import "crypto/rand"

var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ") 

func randomString(len int) string {
	outputBytes := make([]rune, len)
	randBytes := make([]byte, len)
	rand.Read(randBytes)
	for i := range outputBytes {
		j := int(randBytes[i]) % len
		outputBytes[i] = letters[j]
	}
	return string(outputBytes)
}