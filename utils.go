package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/url"
)

func validateParams(params ...string) {
	for _, v := range params {
		if len(v) == 0 {
			panic(fmt.Errorf("ERROR: %s", "BAD PARAMS"))
		}
	}
}

func handleError(err error) {
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}
}

func randToken(n int) string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func generateTempTokenUrl(tempToken string) string {
	var buf bytes.Buffer
	buf.WriteString(BaseUrl)
	v := url.Values{
		"temporary_token": {tempToken},
	}
	buf.WriteString("?")
	buf.WriteString(v.Encode())
	return buf.String()
}
