package iout

type FnWriter func([]byte) (int, error)

func (w FnWriter) Write(p []byte) (int, error) {
	return w(p)
}

type FnReader func([]byte) (int, error)

func (r FnReader) Read(p []byte) (int, error) {
	return r(p)
}

func CopyBytes(b []byte) []byte {
	p := make([]byte, len(b))
	copy(p, b)
	return p
}
