//go:build !windows

package findfont

/*
#cgo pkg-config: fontconfig

#include <stdlib.h>
#include <string.h>
#include <fontconfig/fontconfig.h>

char *find_font(const char *font_name) {
    FcInit();
    FcPattern *pattern = FcNameParse((const FcChar8 *)font_name);
    FcConfigSubstitute(NULL, pattern, FcMatchPattern);
    FcDefaultSubstitute(pattern);

    FcResult result;
    FcPattern *match = FcFontMatch(NULL, pattern, &result);
    FcChar8 *file = NULL;

    if (match) {
        if (FcPatternGetString(match, FC_FILE, 0, &file) == FcResultMatch) {
            char *file_path = strdup((const char *)file);
            FcPatternDestroy(match);
            FcPatternDestroy(pattern);
            FcFini();
            return file_path;
        }
        FcPatternDestroy(match);
    }

    FcPatternDestroy(pattern);
    FcFini();
    return NULL;
}
*/
import "C"
import (
	"errors"
	"os"
	"unsafe"
)

var ErrNoMatch = errors.New("font not found")

func GetFontData(query string) ([]byte, error) {
	cquery := C.CString(query)
	defer C.free(unsafe.Pointer(cquery))

	cpath := C.find_font(cquery)
	if cpath == nil {
		return nil, ErrNoMatch
	}
	defer C.free(unsafe.Pointer(cpath))

	path := C.GoString(cpath)
	return os.ReadFile(path)
}
