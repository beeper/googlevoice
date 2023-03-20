package util

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func ExtractSID(cookie string) string {
	var apiSIDHash string

	for _, cookies := range strings.Split(cookie, "; ") {
		cookieParts := strings.SplitN(cookies, "=", 2)
		if cookieParts[0] == "SAPISID" {
			// Calculate SHA-1 hash
			hasher := sha1.New()
			now := time.Now().Unix()
			hasher.Write([]byte(fmt.Sprintf("%d %s %s", now, cookieParts[1], "https://voice.google.com")))
			shaHash := hex.EncodeToString(hasher.Sum(nil))
			apiSIDHash = fmt.Sprintf("%sHASH %d_%s", cookieParts[0], now, shaHash)
			break
		}
	}

	return apiSIDHash
}
