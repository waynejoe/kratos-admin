package utils

import (
	"crypto/md5" // #nosec G501
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func splitUrl(url string) []string {
	reg := regexp.MustCompile(`^(http://|https://)?([^/?]+)(/[^?]*)?(\\?.*)?$`)
	return reg.FindStringSubmatch(url)
}

func getMd5(text string) string {
	hashByte := md5.Sum([]byte(text)) // #nosec G401
	return hex.EncodeToString(hashByte[:])
}

func GenTypeAUrl(url string, key string, ts int64) string {
	uid := "0"
	signName := "auth_key"
	params := splitUrl(url)
	scheme, host, path, args := params[1], params[2], params[3], params[4]
	randstr := GetRandomString(10)
	text := fmt.Sprintf("%s-%d-%s-%s-%s", path, ts, randstr, uid, key)
	hash := getMd5(text)
	authArg := fmt.Sprintf("%s=%d-%s-%s-%s", signName, ts, randstr, uid, hash)
	if args == "" {
		return fmt.Sprintf("%s%s%s?%s", scheme, host, path, authArg)
	} else {
		return fmt.Sprintf("%s%s%s%s&%s", scheme, host, path, args, authArg)
	}
}

// VerifyTypeAUrl 反验证带有签名的 URL
func VerifyTypeAUrl(signedUrl string, key string, validTs int64) bool {
	// 解析 URL
	u, err := url.Parse(signedUrl)
	if err != nil {
		return false
	}

	// 提取签名参数
	query := u.Query()
	signName := "auth_key"
	authArg := query.Get(signName)
	if authArg == "" {
		return false
	}

	// 拆分签名参数
	parts := strings.Split(authArg, "-")
	if len(parts) != 4 {
		return false
	}
	tsStr := parts[0]
	randStr := parts[1]
	uid := parts[2]
	receivedHash := parts[3]

	// 解析时间戳
	var ts int64
	if ts, err = strconv.ParseInt(tsStr, 10, 64); err != nil {
		return false
	}
	// ts过小，拒绝
	if ts < validTs {
		return false
	}

	// 重新计算哈希值
	path := u.Path
	text := fmt.Sprintf("%s-%d-%s-%s-%s", path, ts, randStr, uid, key)
	expectedHash := getMd5(text)

	// 比较哈希值
	return expectedHash == receivedHash
}

// URLWithCDN 图片cdn地址拼接
func URLWithCDN(path, cdnHost string) string {
	// 兼容空串
	if path == "" {
		return path
	}
	// 完整链接直接返回
	if strings.HasPrefix(path, "http") || strings.HasPrefix(path, "https") {
		return path
	}
	// 拼接
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return cdnHost + path
}

// ParseURI
func GetRelativePath(fullURL string) string {
	// 解析URL
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return ""
	}

	// 组合路径、查询和片段为相对路径
	relativePath := parsedURL.Path
	if parsedURL.RawQuery != "" {
		relativePath += "?" + parsedURL.RawQuery
	}
	if parsedURL.Fragment != "" {
		relativePath += "#" + parsedURL.Fragment
	}

	return relativePath
}
