package randomutil

import (
	"math/rand"
	"strings"
)

func randomString(src string, n int) string {
	var str strings.Builder
	var k = len(src)
	for i := 0; i < n; i++ {
		str.WriteByte(src[rand.Intn(k)])
	}
	return str.String()
}

func RandomAlphaNumString(n int) string {
	const s = "abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return randomString(s, n)
}

func RandomNumString(n int) string {
	const s = "0123456789"
	return randomString(s, n)
}
