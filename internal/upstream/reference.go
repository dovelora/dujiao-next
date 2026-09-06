package upstream

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// Reference 是上游系统的 opaque 字符串标识。兼容读取旧节点返回的 JSON 数字，
// 对外始终编码为字符串，避免 JavaScript 对大整数引用造成精度损失。
type Reference string

func (r Reference) String() string { return string(r) }

func (r Reference) IsZero() bool {
	value := strings.TrimSpace(string(r))
	return value == "" || value == "0"
}

func (r Reference) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.TrimSpace(string(r)))
}

func (r *Reference) UnmarshalJSON(data []byte) error {
	if r == nil {
		return errors.New("upstream reference is nil")
	}
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		*r = ""
		return nil
	}
	if len(data) > 0 && data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*r = Reference(strings.TrimSpace(value))
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return err
	}
	parsed, err := strconv.ParseUint(number.String(), 10, 64)
	if err != nil {
		return errors.New("upstream numeric reference must be an unsigned integer")
	}
	*r = Reference(strconv.FormatUint(parsed, 10))
	return nil
}
