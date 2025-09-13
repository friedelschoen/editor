package fontutil

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/image/font/opentype"
)

func FontMatch(fontstr string) (string, *opentype.FaceOptions, error) {
	var conf opentype.FaceOptions

	var buf strings.Builder
	cmd := exec.Command("fc-match", "-f", "%{size} %{dpi} %{file}", fontstr)
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", nil, err
	}
	line := strings.TrimSpace(buf.String())
	fields := strings.SplitN(line, " ", 3)
	conf.Size, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return "", nil, err
	}
	conf.DPI, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return "", nil, err
	}

	return fields[2], &conf, nil
}
