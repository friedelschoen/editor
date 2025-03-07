// Source code editor in pure Go.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/friedelschoen/editor/internal/core"
	"github.com/friedelschoen/editor/internal/lsproto"

	// imports that can't be imported from core (cyclic import)
	_ "github.com/friedelschoen/editor/internal/contentcmds"
	_ "github.com/friedelschoen/editor/internal/internalcmds"
)

func main() {
	opt := &core.Options{}

	// flags
	flag.StringVar(&opt.Font, "font", "regular", "font: regular, medium, mono, or a filename")
	flag.Float64Var(&opt.FontSize, "fontsize", 12, "")
	flag.StringVar(&opt.FontHinting, "fonthinting", "full", "font hinting: none, vertical, full")
	flag.Float64Var(&opt.DPI, "dpi", 72, "monitor dots per inch")
	flag.IntVar(&opt.TabWidth, "tabwidth", 8, "")
	// flag.StringVar(&opt.CarriageReturnRune, "carriagereturnrune", "", "replacement rune for carriage return")
	flag.StringVar(&opt.WrapLineRune, "wraplinerune", "←", "code for wrap line rune, can be set to zero")
	flag.StringVar(&opt.ColorTheme, "colortheme", "light", "color theme")
	flag.IntVar(&opt.ScrollBarWidth, "scrollbarwidth", 0, "Textarea scrollbar width in pixels. A value of 0 takes 3/4 of the font size.")
	flag.BoolVar(&opt.ScrollBarLeft, "scrollbarleft", true, "set scrollbars on the left side")
	flag.BoolVar(&opt.Shadows, "shadow.s", true, "shadow effects on some elements")
	flag.StringVar(&opt.SessionName, "sn", "", "open existing session")
	flag.StringVar(&opt.SessionName, "sessionname", "", "open existing session")
	flag.BoolVar(&opt.UseMultiKey, "usemultikey", false, "use multi-key to compose characters (Ex: [multi-key, ~, a] = ã)")
	flag.StringVar(&opt.Plugins, "plugins", "", "comma separated string of plugin filenames")
	flag.Var(&opt.LSProtos, "lsproto", "Language-server-protocol register options. Can be specified multiple times.\nFormat: language,fileExtensions,network{tcp|tcpclient|stdio},command,optional{stderr,nogotoimpl}\nFormat notes:\n\tif network is tcp, the command runs in a template with vars: {{.Addr}}.\n\tif network is tcpclient, the command should be an ipaddress.\nExamples:\n\t"+strings.Join(lsproto.RegistrationExamples(), "\n\t"))
	flag.Var(&opt.PreSaveHooks, "presavehook", "Run program before saving a file. Uses stdin/stdout. Can be specified multiple times. By default, a \"goimports\" entry is auto added if no entry is defined for the \"go\" language.\nFormat: language,fileExtensions,cmd\nExamples:\n"+
		"\tgo,.go,goimports\n"+
		"\tcpp,\".cpp .hpp\",\"\\\"clang-format --style={'opt1':1,'opt2':2}\\\"\"\n"+
		"\tpython,.py,python_formatter")
	flag.BoolVar(&opt.ZipSessionsFile, "zipsessionsfile", false, "Save sessions in a zip. Useful for 100+ sessions. Does not delete the plain file. Beware that the file might not be easily editable as in a plain file.")
	cpuProfileFlag := flag.String("cpuprofile", "", "profile cpu filename")
	version := flag.Bool("version", false, "output version and exit")

	flag.Parse()

	if err := core.ParseConfig(opt, core.ConfigFilename()); err != nil {
		log.Println(err)
		return
	}

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

	if err := core.RunEditor(opt); err != nil {
		log.Println(err) // fatal
		os.Exit(1)
	}
}
