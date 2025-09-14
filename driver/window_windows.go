//go:build windows && !xproto

package driver

import (
	"github.com/friedelschoen/editor/driver/windriver"
)

func NewWindow() (Window, error) {
	return windriver.NewWindow()
}
