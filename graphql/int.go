package graphql

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

func MarshalInt(i int) Marshaler {
	return WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Itoa(i))
	})
}

func UnmarshalInt(v any) (int, error) {
	switch v := v.(type) {
	case string:
		return strconv.Atoi(v)
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case json.Number:
		return strconv.Atoi(string(v))
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("%T is not an int", v)
	}
}

func MarshalInt64(i int64) Marshaler {
	return WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.FormatInt(i, 10))
	})
}

func UnmarshalInt64(v any) (int64, error) {
	switch v := v.(type) {
	case string:
		return strconv.ParseInt(v, 10, 64)
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case json.Number:
		return strconv.ParseInt(string(v), 10, 64)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("%T is not an int", v)
	}
}

func MarshalInt32(i int32) Marshaler {
	return WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.FormatInt(int64(i), 10))
	})
}

func UnmarshalInt32(v any) (int32, error) {
	switch v := v.(type) {
	case string:
		iv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return safeCastInt32(iv)
	case int:
		return safeCastInt32(int64(v))
	case int64:
		return safeCastInt32(v)
	case json.Number:
		iv, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0, err
		}
		return safeCastInt32(iv)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("%T is not an int", v)
	}
}

type Int32OverflowError struct {
	IVal int64
	SVal string
}

func (e *Int32OverflowError) Error() string {
	if e.SVal != "" {
		return fmt.Sprintf("%s overflows int32", e.SVal)
	}
	return fmt.Sprintf("%d overflows int32", e.IVal)
}

func safeCastInt32(i int64) (int32, error) {
	if i > 2147483647 || i < -2147483648 {
		return 0, &Int32OverflowError{IVal: i}
	}
	return int32(i), nil
}
