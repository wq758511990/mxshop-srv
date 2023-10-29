package main

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"io"
	"mxshop_srvs/user_srv/global"
	"mxshop_srvs/user_srv/model"
)

func genMd5(code string) string {
	ans := md5.New()
	_, _ = io.WriteString(ans, code)
	return hex.EncodeToString(ans.Sum(nil))
}
func main() {
	// 自定义options
	options := &password.Options{SaltLen: 16, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}
	salt, encodedPwd := password.Encode("generic password", options)
	pwd := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)

	for i := 0; i < 5; i++ {
		user := model.User{
			NickName: fmt.Sprintf("bobby%d", i),
			Mobile:   fmt.Sprintf("18766664444-%d", i),
			Password: pwd,
		}
		global.DB.Save(&user)
	}
}
