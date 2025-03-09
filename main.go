package main

import (
	"context"
	"fmt"
	"os"

	"github.com/techarm/jclog/cmd"
)

const Version = "0.3.2"

func main() {
	err := cmd.NewRootCommand().Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
