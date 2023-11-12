package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFormat = errors.New("invalid minutes format")

type Mins int32

func (m Mins) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", m)
	quotedJSONValue := strconv.Quote(jsonValue)
	return []byte(quotedJSONValue), nil
}

func (m *Mins) UnmarshalJSON(jsonValue []byte) error {

	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	fmt.Printf("Unquoted JSON value: %s\n", unquotedJSONValue)

	parts := strings.Split(unquotedJSONValue, " ")
	fmt.Printf("Parts after splitting unquoted JSON value: %v\n", parts)

	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		fmt.Printf("Error parsing integer from parts: %v\n", err)

		return ErrInvalidRuntimeFormat
	}
	*m = Mins(i)
	return nil
}
