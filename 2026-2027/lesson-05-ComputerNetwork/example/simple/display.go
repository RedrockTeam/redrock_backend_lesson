package simple

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// display 串行化 stdout 的写入，让提示符始终停留在最后一行，
// 收到的消息打印在提示符上方。
type display struct {
	mu      sync.Mutex
	prompt  string
	visible bool
	closed  bool
	tty     bool
}

func newDisplay(prompt string) *display {
	return &display{
		prompt: prompt,
		tty:    isTerminal(os.Stdout),
	}
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// ShowPrompt 在提示符尚未显示时打印它。
func (d *display) ShowPrompt() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed || d.visible {
		return
	}
	fmt.Fprint(os.Stdout, d.prompt)
	d.visible = true
}

// ConsumePrompt 表示用户已按下回车，光标已经移到下一行，
// 提示符不再位于屏幕上。
func (d *display) ConsumePrompt() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}
	// 清除这期间读协程可能重绘出来的提示符。
	if d.visible && d.tty {
		fmt.Fprint(os.Stdout, "\r\033[2K")
	}
	d.visible = false
}

// PrintLine 在提示符上方打印一行消息，然后把提示符重绘到最底部。
func (d *display) PrintLine(msg string) {
	msg = strings.TrimRight(msg, "\r\n")

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}
	if d.visible && d.tty {
		fmt.Fprint(os.Stdout, "\r\033[2K")
	}
	fmt.Fprintln(os.Stdout, msg)
	if d.visible && d.tty {
		fmt.Fprint(os.Stdout, d.prompt)
	}
}

// Close 在退出前把光标移到提示符的下一行。
func (d *display) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}
	if d.visible {
		fmt.Fprintln(os.Stdout)
		d.visible = false
	}
	d.closed = true
}
