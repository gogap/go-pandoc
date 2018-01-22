package server

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"html"
	"text/template"

	"github.com/spf13/cast"
)

var (
	funcMap = template.FuncMap{
		"base64Encode": base64Encode,
		"base64Decode": base64Decode,
		"jsonify":      jsonify,
		"md5":          md5String,
		"toBytes":      toBytes,
		"htmlEscape":   htmlEscape,
		"htmlUnescape": htmlUnescape,
	}
)

func toBytes(content interface{}) ([]byte, error) {

	switch v := content.(type) {
	case []byte:
		{
			return v, nil
		}
	case string:
		{
			return []byte(v), nil
		}
	default:
		{
			str, err := cast.ToStringE(content)
			if err != nil {
				return nil, err
			}

			return []byte(str), nil
		}
	}
}

func base64Decode(content interface{}) (string, error) {
	conv, err := cast.ToStringE(content)
	if err != nil {
		return "", err
	}

	dec, err := base64.StdEncoding.DecodeString(conv)
	return string(dec), err
}

func base64Encode(content interface{}) (string, error) {
	conv, err := cast.ToStringE(content)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString([]byte(conv)), nil
}

func md5String(f string) string {
	h := md5.New()
	h.Write([]byte(f))
	return hex.EncodeToString(h.Sum([]byte{}))
}

func jsonify(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func htmlEscape(s interface{}) (string, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", err
	}

	return html.EscapeString(ss), nil
}

func htmlUnescape(s interface{}) (string, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", err
	}

	return html.UnescapeString(ss), nil
}
