package main

import (
	"context"
	"fmt"
	"os"

	"github.com/techarm/json-log-viewer/cmd"
)

func main() {
	err := cmd.NewRootCommand().Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
