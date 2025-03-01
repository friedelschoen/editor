package osutil

import (
	"fmt"
	"strings"

	"github.com/friedelschoen/glake/util/strconvutil"
)

func GetEnv(env []string, key string) string {
	for i := len(env) - 1; i >= 0; i-- { // last entry has precedence
		s := env[i]
		k, v, ok := splitEnvVar(s)
		if !ok {
			continue
		}
		if k == key {
			return v
		}
	}
	return ""
}

func UnquoteEnvValues(env []string) []string {
	w := []string{}
	for _, s := range env {
		k, v, ok := splitEnvVar(s)
		if !ok {
			continue
		}
		// NOTE: strconv.Unquote() fails on singlequotes with len>6 runes
		if v2, ok := strconvutil.BasicUnquote(v); ok {
			w = append(w, keyValStr(k, v2))
		} else {
			w = append(w, s)
		}
	}
	return w
}

func keyValStr(key, value string) string {
	return fmt.Sprintf("%v=%v", key, value)
}

func splitEnvVar(s string) (string, string, bool) {
	return strings.Cut(s, "=")
}
