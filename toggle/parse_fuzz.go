package toggle

import "bytes"

/*
Build: go-fuzz-build github.com/globusdigital/feature-toggles/toggle

Run: go-fuzz -bin=./toggle-fuzz.zip -workdir fuzz-data
*/

func Fuzz(in []byte) int {
	_, err := ParseCondition(bytes.NewReader(in))
	if err == nil {
		return 1
	}

	return 0
}
