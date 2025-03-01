package goutil

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/friedelschoen/glake/util/osutil"
)

func GoEnv(dir string) []string {
	w, err := GoEnv2(dir)
	if err != nil {
		return nil
	}
	return w
}
func GoEnv2(dir string) ([]string, error) {
	// not the same as os.Environ which has entries like PATH

	c := osutil.NewCmdI2([]string{"go", "env"})
	c = osutil.NewShellCmd(c, false)
	c.Cmd().Dir = dir
	bout, err := osutil.RunCmdICombineStderrErr(c)
	if err != nil {
		return nil, err
	}
	env := strings.Split(string(bout), "\n")

	// clear "set " prefix
	if runtime.GOOS == "windows" {
		for i, s := range env {
			env[i] = strings.TrimPrefix(s, "set ")
		}
	}

	env = osutil.UnquoteEnvValues(env)

	return env, nil
}

func GoRoot() string {
	// doesn't work well in windows
	//return runtime.GOROOT()

	return GetGoRoot(GoEnv(""))
}

func GoPath() []string {
	return GetGoPath(GoEnv(""))
}

func GetGoRoot(env []string) string {
	return osutil.GetEnv(env, "GOROOT")
}

func GetGoPath(env []string) []string {
	//res := []string{}
	//a := osutil.GetEnv(env, "GOPATH")
	//if a != "" {
	//	res = append(res, filepath.SplitList(a)...)
	//} else {
	//	// from go/build/build.go:274
	//	res = append(res, filepath.Join(osutil.HomeEnvVar(), "go"))
	//}
	//return res

	a := osutil.GetEnv(env, "GOPATH")
	return filepath.SplitList(a)
}

// returns version as in "1.0" without the "go" prefix
