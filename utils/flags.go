package utils

import (
	"fmt"
	"strings"
)

type StringSliceFlag []string

func (i *StringSliceFlag) String() string {
	return ""
}

func (i *StringSliceFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func MakeHAProxyParams(flags StringSliceFlag) (HAProxyParams, error) {
	params := HAProxyParams{
		Defaults: map[string][]string{},
		Globals:  map[string][]string{},
	}

	for _, flag := range flags {
		parts := strings.Split(flag, "=")
		if len(parts) != 2 {
			return params, fmt.Errorf("bad haproxy-param flag %s, expected {type}.{name}={value}", flag)
		}

		dot := strings.IndexByte(parts[0], '.')
		if dot == -1 {
			return params, fmt.Errorf("bad haproxy-param flag %s, expected {type}.{name}={value}", flag)
		}

		var t map[string][]string
		switch parts[0][:dot] {
		case "defaults":
			t = params.Defaults
		case "global":
			t = params.Globals
		default:
			return params, fmt.Errorf("bad haproxy-param flag %s, param type must be `defaults` or `global`", flag)
		}

		_, ok := t[parts[0][dot+1:]]
		if !ok  {
			t[parts[0][dot+1:]] = []string{}
		}
		t[parts[0][dot+1:]] = append(t[parts[0][dot+1:]], parts[1])
	}

	return params, nil
}
