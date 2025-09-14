package wimage

import (
	"image"
	"unsafe"

	"github.com/friedelschoen/editor/util/imageutil"
)

type ShmImgWrap struct {
	Img   *imageutil.BGRA
	shmId uintptr
	addr  uintptr
}

func NewShmImgWrap(r image.Rectangle) (*ShmImgWrap, error) {
	size := imageutil.BGRASize(&r)
	shmId, addr, err := ShmOpen(size)
	if err != nil {
		return nil, err
	}

	// mask shared mem into a slice
	buf := unsafe.Slice((*byte)(unsafe.Pointer(addr)), size) // FIXME: warning: possible misuse of unsafe.Pointer?

	img := imageutil.NewBGRAFromBuffer(buf, &r)
	imgWrap := &ShmImgWrap{Img: img, shmId: shmId, addr: addr}
	return imgWrap, nil
}

func (imgWrap *ShmImgWrap) Close() error {
	return ShmClose(imgWrap.shmId, imgWrap.addr)
}
