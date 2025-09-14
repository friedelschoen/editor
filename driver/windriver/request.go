package windriver

import (
	"image"
	"image/draw"

	"github.com/jmigpin/editor/util/uiutil/event"
)

type Request any

type ReqClose struct{}
type ReqWindowSetName struct{ Name string }
type ReqImage struct{ ReplyImg draw.Image }
type ReqImagePut struct{ Rect image.Rectangle }
type ReqImageResize struct{ Rect image.Rectangle }
type ReqCursorSet struct{ Cursor event.Cursor }
type ReqPointerQuery struct{ ReplyP image.Point }
type ReqPointerWarp struct{ P image.Point }
type ReqClipboardDataGet struct{ ReplyS string }
type ReqClipboardDataSet struct{ Str string }
