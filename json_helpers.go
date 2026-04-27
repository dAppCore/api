// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"strconv"

	core "dappco.re/go/core"
)

type jsonNumber string

func (n jsonNumber) String() string {
	return string(n)
}

func (n jsonNumber) Float64() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

func (n jsonNumber) Int64() (int64, error) {
	return strconv.ParseInt(string(n), 10, 64)
}

func (n jsonNumber) MarshalJSON() ([]byte, error) {
	if n == "" {
		return nil, core.E("jsonNumber.MarshalJSON", "empty JSON number", nil)
	}
	return []byte(n), nil
}

type jsonRawMessage []byte

func (m jsonRawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return append([]byte(nil), m...), nil
}

func (m *jsonRawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return core.E("jsonRawMessage.UnmarshalJSON", "target is nil", nil)
	}
	*m = append((*m)[:0], data...)
	return nil
}

type jsonValue struct {
	value any
}

func (v *jsonValue) UnmarshalJSON(data []byte) error {
	if v == nil {
		return core.E("jsonValue.UnmarshalJSON", "target is nil", nil)
	}

	text := core.Trim(string(data))
	if text == "" {
		return core.E("jsonValue.UnmarshalJSON", "empty JSON value", nil)
	}

	switch text[0] {
	case '{':
		var raw map[string]jsonValue
		if err := unmarshalCoreJSON(data, &raw); err != nil {
			return err
		}
		out := make(map[string]any, len(raw))
		for key, item := range raw {
			out[key] = item.value
		}
		v.value = out
	case '[':
		var raw []jsonValue
		if err := unmarshalCoreJSON(data, &raw); err != nil {
			return err
		}
		out := make([]any, len(raw))
		for i, item := range raw {
			out[i] = item.value
		}
		v.value = out
	case '"':
		var out string
		if err := unmarshalCoreJSON(data, &out); err != nil {
			return err
		}
		v.value = out
	case 't', 'f':
		var out bool
		if err := unmarshalCoreJSON(data, &out); err != nil {
			return err
		}
		v.value = out
	case 'n':
		var out any
		if err := unmarshalCoreJSON(data, &out); err != nil {
			return err
		}
		v.value = out
	default:
		v.value = jsonNumber(text)
	}

	return nil
}

func decodeJSONValuePreserveNumbers(data []byte) (any, error) {
	var out jsonValue
	if err := unmarshalCoreJSON(data, &out); err != nil {
		return nil, err
	}
	return out.value, nil
}

func marshalCoreJSON(value any) ([]byte, error) {
	result := core.JSONMarshal(value)
	if !result.OK {
		return nil, coreResultError(result)
	}

	data, ok := result.Value.([]byte)
	if !ok {
		return nil, core.E("marshalCoreJSON", "encoded JSON was not bytes", nil)
	}
	return data, nil
}

func marshalCoreJSONIndent(value any, prefix, indent string) ([]byte, error) {
	data, err := marshalCoreJSON(value)
	if err != nil {
		return nil, err
	}
	return indentJSON(data, prefix, indent), nil
}

func indentJSON(data []byte, prefix, indent string) []byte {
	out := core.NewBuffer()
	level := 0
	inString := false
	escaped := false

	writeIndent := func() {
		out.WriteString(prefix)
		for i := 0; i < level; i++ {
			out.WriteString(indent)
		}
	}

	for i := 0; i < len(data); i++ {
		b := data[i]
		if inString {
			out.WriteByte(b)
			if escaped {
				escaped = false
				continue
			}
			switch b {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}
			continue
		}

		switch b {
		case '"':
			inString = true
			out.WriteByte(b)
		case '{', '[':
			out.WriteByte(b)
			if i+1 < len(data) && ((b == '{' && data[i+1] == '}') || (b == '[' && data[i+1] == ']')) {
				i++
				out.WriteByte(data[i])
				continue
			}
			level++
			out.WriteByte('\n')
			writeIndent()
		case '}', ']':
			level--
			out.WriteByte('\n')
			writeIndent()
			out.WriteByte(b)
		case ',':
			out.WriteByte(b)
			out.WriteByte('\n')
			writeIndent()
		case ':':
			out.WriteByte(b)
			out.WriteByte(' ')
		default:
			out.WriteByte(b)
		}
	}

	return out.Bytes()
}

func unmarshalCoreJSON(data []byte, target any) error {
	result := core.JSONUnmarshal(data, target)
	if result.OK {
		return nil
	}
	return coreResultError(result)
}
