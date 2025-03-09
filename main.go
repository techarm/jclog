package main

import (
	"context"
	"fmt"
	"os"

	"github.com/techarm/jclog/cmd"
)

func main() {
	err := cmd.NewRootCommand().Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
