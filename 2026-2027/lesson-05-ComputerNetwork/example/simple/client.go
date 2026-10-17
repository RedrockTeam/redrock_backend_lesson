package simple

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func Client(addr string) {
	// 全局连接
	var Conn net.Conn

	// 初始化全局Context
	ctx, cancel := context.WithCancel(context.Background())

	// 终端显示
	d := newDisplay("[Chat Client]>>> ")
	defer d.Close()

	// 初始化连接
	Conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	// 读携程
	go func() {
		reader := bufio.NewReader(Conn)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := reader.ReadString('\n')
				if err != nil {
					if errors.Is(err, io.EOF) {
						d.PrintLine("连接已关闭")
					} else {
						d.PrintLine(fmt.Sprintf("[Read Error]: %v", err))
					}
					cancel()
					return
				}
				d.PrintLine(fmt.Sprintf("[%v]: %v", Conn.RemoteAddr(), msg))
			}
		}
	}()

	// 写循环
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return

		default:
			d.ShowPrompt()
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					d.PrintLine(fmt.Sprintf("[Scan Error]: %v", err))
				}
				cancel()
				return
			}
			d.ConsumePrompt()

			msg := scanner.Text()

			if msg == "exit" {
				cancel()
			}

			_, err = fmt.Fprintln(Conn, msg)
			if err != nil {
				d.PrintLine(fmt.Sprintf("[Write Error]: %v", err))
				continue
			}
		}
	}
}
