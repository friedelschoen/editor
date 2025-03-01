package osutil

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func HomeEnvVar() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

// Result does not start with separator.

func FilepathClean(s string) string {
	return filepath.Clean(s)
}

func GetFreeTcpPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	p := l.Addr().(*net.TCPAddr).Port
	return p, nil
}

// doesn't wait for the cmd to end
func OpenBrowser(url string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		c = exec.Command("open", url)
	default: // linux, others...
		c = exec.Command("xdg-open", url)
	}
	return cmdStartWaitAsync(c)
}

// doesn't wait for the cmd to end
func OpenExternal(name string) error {
	c := (*exec.Cmd)(nil)
	switch runtime.GOOS {
	case "windows":
		// TODO: review
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", name)
	case "darwin":
		c = exec.Command("open", name)
	default: // linux, others...
		c = exec.Command("xdg-open", name)
	}
	return cmdStartWaitAsync(c)
}

// doesn't wait for the cmd to end
func OpenFilemanager(filename string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		c = exec.Command("explorer", "/select,"+filename)
	case "darwin":
		c = exec.Command("open", filename)
	default: // linux, others...
		c = exec.Command("xdg-open", filename)
	}
	return cmdStartWaitAsync(c)
}

// doesn't wait for the cmd to end
func OpenTerminal(filename string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		return errors.New("todo")
	case "darwin":
		// TODO: review
		c = exec.Command("terminal", filename)
	default: // linux, others...
		c = exec.Command("x-terminal-emulator", "--working-directory="+filename)
	}
	return cmdStartWaitAsync(c)
}

func cmdStartWaitAsync(c *exec.Cmd) error {
	if err := c.Start(); err != nil {
		return err
	}
	go c.Wait() // async to let run, but wait to clear resources
	return nil
}
