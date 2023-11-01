package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

func genMd5(code string) string {
	ans := md5.New()
	_, _ = io.WriteString(ans, code)
	return hex.EncodeToString(ans.Sum(nil))
}
