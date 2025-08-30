package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

type HMAC struct {
	hmac hash.Hash
}

func NewHMAC(key string) HMAC {
	h := hmac.New(sha256.New, []byte(key))
	return HMAC{
		hmac: h,
	}
}

func (h HMAC) Hash(input string) string {
	h.hmac.Reset()                              // reset to start fresh
	h.hmac.Write([]byte(input))                 // write data to be hashed
	b := h.hmac.Sum(nil)                        // compute hash
	return base64.URLEncoding.EncodeToString(b) // encode to string
}
