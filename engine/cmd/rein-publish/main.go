// rein-publish — the registry generator CLI. Logic lives in
// engine/internal/publish; see spec/registry.md.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joaofreitas04/freerein/engine/internal/publish"
)

func main() {
	out := flag.String("out", "", "registry output directory (holds index.json and archives)")
	flag.Parse()
	if *out == "" || flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: rein-publish --out <registry-dir> <component-dir>...")
		os.Exit(2)
	}
	if err := publish.Run(*out, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, "rein-publish:", err)
		os.Exit(1)
	}
}
