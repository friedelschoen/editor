// Source code editor in pure Go.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime/pprof"

	"github.com/jmigpin/editor/core"

	// imports that can't be imported from core (cyclic import)
	_ "github.com/jmigpin/editor/core/contentcmds"
	_ "github.com/jmigpin/editor/core/internalcmds"
)

func configPath() string {
	confdir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "/"
		}
		confdir = path.Join(home, ".config")
	}
	return path.Join(confdir, "editor.json")
}

func main() {
	opt := core.Options{
		Font:               "monospace",
		ColorTheme:         "light",
		TabWidth:           8,
		CarriageReturnRune: "␍",
		WrapLineRune:       "←",
	}

	if conffile, err := os.ReadFile(configPath()); err == nil {
		json.Unmarshal(conffile, &opt)
	}

	cpuProfileFlag := flag.String("cpuprofile", "", "profile cpu filename")
	version := flag.Bool("version", false, "output version and exit")

	flag.Parse()
	opt.Filenames = flag.Args()

	log.SetFlags(log.Lshortfile)

	if *version {
		fmt.Printf("version: %v\n", core.Version())
		return
	}

	if *cpuProfileFlag != "" {
		f, err := os.Create(*cpuProfileFlag)
		if err != nil {
			log.Println(err)
			return
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if err := core.RunEditor(&opt); err != nil {
		log.Println(err) // fatal
		os.Exit(1)
	}
}
