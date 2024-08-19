package main

import (
	"github.com/0x0FACED/link-saver-api/internal/server"
)

func main() {
	if err := server.Start(); err != nil {
		panic("cant start server: " + err.Error())
	}
}
