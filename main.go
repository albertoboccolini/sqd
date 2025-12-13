package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/albertoboccolini/sqd/services"
)

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("Usage: sqd 'SELECT * FROM file.txt WHERE content LIKE pattern'")
		os.Exit(1)
	}

	sql := strings.Join(flag.Args(), " ")
	cmd := services.ParseSQL(sql)

	files := services.FindFiles(cmd.File)
	if len(files) == 0 {
		fmt.Println("No files found")
		os.Exit(1)
	}

	services.ExecuteCommand(cmd, files)
}
