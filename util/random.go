package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// This function will automatic execute when this package is first used
func init() {
	//Make sure whem every time run the code the generated value will be difference
	rand.Seed(time.Now().UnixNano())
}

//RandomInt generate a random interger between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

//RandomString generate a ramdom string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

//RandomName generate random name
func RandomName() string {
	return RandomString(6)
}

//RandomDob generate random date of birth
func RandomDob() string {
	return fmt.Sprintf("August %d, %d", RandomInt(1, 30), RandomInt(1000, 2000))
}

//RandomLink generate a random bio link
func RandomBioLink() string {
	bioLinks := []string{
		"https://en.wikipedia.org/wiki/Elon_Musk",
		"https://en.wikipedia.org/wiki/Jeff_Bezos"}
	n := len(bioLinks)
	return bioLinks[rand.Intn(n)]
}

//RandomLink generate a random image link
func RandomImageLink() string {
	imagesLinks := []string{
		"https://upload.wikimedia.org/wikipedia/commons/e/ed/Elon_Musk_Royal_Society.jpg",
		"https://pbs.twimg.com/profile_images/669103856106668033/UF3cgUk4_400x400.jpg"}
	n := len(imagesLinks)
	return imagesLinks[rand.Intn(n)]
}

//RandomEmail generate a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
