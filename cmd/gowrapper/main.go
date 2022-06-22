package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/coc1961/gowrapper/internal/wrapper"
)

func main() {
	src := flag.String("s", "", "source go file")
	name := flag.String("n", "", "interface name")
	addPackage := flag.Bool("p", false, "(Optional) add package name in local objects")

	flag.Parse()

	if *src == "" || *name == "" {
		flag.CommandLine.Usage()
		return
	}
	fileExists := func(filename string) bool {
		info, err := os.Stat(filename)
		if os.IsNotExist(err) {
			return false
		}
		return !info.IsDir()
	}

	if !fileExists(*src) {
		fmt.Fprintf(os.Stderr, "file not found %v", *src)
		return
	}

	mm := wrapper.MockMaker{}
	x := mm.CreateMock(*src, *name, *addPackage)
	fmt.Fprint(os.Stdout, x.String())
}
