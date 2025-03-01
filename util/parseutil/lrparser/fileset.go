package lrparser

// TODO: multiple files (working for single file only)
type FileSet struct {
	Src      []byte // currently, just a single src
	Filename string // for errors only
}

//func (fset *FileSet) SliceFrom(i int) []byte {
//	// TODO: implemented for single file only (need node arg?)
//	return fset.src[i:]
//}
//func (fset *FileSet) SliceTo(i int) []byte {
//	// TODO: implemented for single file only (need node arg?)
//	return fset.src[:i]
//}
