package common

import "github.com/bytedance/sonic"

func StructToString(v interface{}) string {
	str, _ := sonic.Marshal(v) //nolint
	return string(str)
}
