package netstring

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// Marshal stringifies interface, first by looking to see if
// it implements `Stringer` and using that if so,
// then by `fmt.Sprintf("%v", v)`
func Marshal(v interface{}) ([]byte, error) {
	var str string
	if s, ok := v.(string); ok {
		str = s
	} else if s, ok := v.(fmt.Stringer); ok {
		str = s.String()
	} else {
		str = fmt.Sprintf("%v", v)
	}

	length := len([]byte(str))
	return []byte(fmt.Sprintf("%d:%s,", length, str)), nil
}

// Unmarshal unmarshals netstrings into []string
// v *must* by `*[]string`
func Unmarshal(data []byte, v interface{}) error {
	if _, ok := v.(*[]string); ok {
		ary := []string{}
		var err error

		lengthBuf := []byte{}
		bodyBuf := []byte{}
		length := 0

		state := 0x0 // 0x0 == parsing length; 0x1 == reading body
		for i := 0; i < len(data); i++ {
			c := data[i]
			switch state {
			case 0x0:
				if c == ':' {
					length, err = strconv.Atoi(string(lengthBuf))
					if err != nil {
						return err
					}
					state = 1
					lengthBuf = []byte{}
					continue
				}
				if 48 <= c && c <= 57 {
					lengthBuf = append(lengthBuf, c)
					continue
				}
				return fmt.Errorf("expected digit or ':' got %s", string(c))
			case 0x1:
				if length == 0 {
					if c != ',' {
						return fmt.Errorf("expected , got %s", string(c))
					}
					ary = append(ary, string(bodyBuf))
					state = 0x0
					bodyBuf = []byte{}
					continue
				}
				bodyBuf = append(bodyBuf, c)
				length--
			}
		}
		if len(bodyBuf) != 0 || len(lengthBuf) != 0 {
			return fmt.Errorf("extra text at end: %s", string(bodyBuf))
		}
		if length != 0 {
			return fmt.Errorf("didn't finish reading string")
		}

		reflect.ValueOf(v).Elem().Set(reflect.ValueOf(ary))
		return nil
	}
	return errors.New("only accepts `*[]string` for target")
}
