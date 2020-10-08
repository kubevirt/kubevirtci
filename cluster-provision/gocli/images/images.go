package images

import (
	"encoding/base64"
	"encoding/json"
)

var (
	SUFFIXES         = ""
	SuffixByProvider map[string]string
)

func init() {
	if len(SUFFIXES) > 0 {
		suffixes, err := base64.StdEncoding.DecodeString(SUFFIXES)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal([]byte(suffixes), &SuffixByProvider)
		if err != nil {
			panic(err)
		}
	}
}
