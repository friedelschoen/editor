package core

import (
	"encoding/json"
	"errors"
	"os"
	"path"
)

func ConfigFilename() string {
	cdir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return path.Join(cdir, "goeditor", "config.json")
}

func ParseConfig(opt *Options, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	err = dec.Decode(opt)
	if err != nil {
		return err
	}
	if dec.More() {
		return errors.New("more data after json-string")
	}

	return nil
}
