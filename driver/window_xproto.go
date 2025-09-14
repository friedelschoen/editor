//go:build !windows || (windows && xproto)

package driver

import "github.com/friedelschoen/editor/driver/xdriver"

func NewWindow() (Window, error) {
	return xdriver.NewWindow()
}
