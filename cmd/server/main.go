package main

import (
	"flag"
	"fmt"
)

var (
	Name     = "app"
	Version  = "dev"
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "configs/config.yaml", "config path")
}

func main() {
	flag.Parse()
	fmt.Printf("%s %s started with config %s\n", Name, Version, flagconf)
	fmt.Println("This is a generated Kernel service skeleton. Add proto contracts under api/ and run make api.")
}
