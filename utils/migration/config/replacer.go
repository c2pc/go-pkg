package config

import (
	"crypto/rand"
	"net"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
)

var Replacer = map[string]func() string{
	"{{ip_address}}":   ipReplacer,
	"{{random}}":       randomReplacer(16),
	"{{random8}}":      randomReplacer(8),
	"{{random32}}":     randomReplacer(32),
	"{{random64}}":     randomReplacer(64),
	"{{project_name}}": projectNameReplacer,
}

func replace(s string) string {
	newS := s
	for k, v := range Replacer {
		index := strings.Index(strings.ToLower(newS), strings.ToLower(k))
		if index != -1 {
			newS = strings.Replace(newS, k, v(), -1)
		}
	}

	return newS
}

func ipReplacer() string {
	tt, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}
	for _, t := range tt {
		aa, err := t.Addrs()
		if err != nil {
			return "localhost"
		}
		for _, a := range aa {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}

			v4 := ipnet.IP.To4()
			if v4 == nil || v4[0] == 127 {
				continue
			}
			return v4.String()
		}
	}
	return "localhost"
}

func randomReplacer(n int) func() string {
	return func() string {
		const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!@()_+-=."
		const letters2 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

		bytes := make([]byte, n)
		_, err := rand.Read(bytes)
		if err != nil {
			return ""
		}

		for i, b := range bytes {
			if i == 0 || i == len(bytes)-1 {
				bytes[i] = letters2[b%byte(len(letters2))]
			} else {
				bytes[i] = letters[b%byte(len(letters))]
			}
		}

		return string(bytes)
	}
}

func projectNameReplacer() string {
	m := regexp.MustCompile("^([a-zA-Z0-9]+)(.*)")
	template := "${1}"

	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return m.ReplaceAllString(os.Args[0], template)
	}

	return m.ReplaceAllString(path.Base(path.Dir(filename)), template)
}
