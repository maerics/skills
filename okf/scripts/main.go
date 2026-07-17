package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	strict := flag.Bool("strict", false, "also fail (nonzero exit) when warnings are present")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-strict] [bundle-path]\n\n"+
			"Validate an OKF knowledge bundle against SPEC.md and the okf skill's\n"+
			"house directives. bundle-path defaults to \".\"; if it contains a .okf\n"+
			"subdirectory, that is used as the bundle root.\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	path := "."
	if flag.NArg() > 0 {
		path = flag.Arg(0)
	}

	root, err := resolveBundleRoot(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "okf-validate:", err)
		os.Exit(1)
	}

	result, err := validateBundle(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "okf-validate:", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "okf-validate:", err)
		os.Exit(1)
	}
	out = append(out, '\n')

	printResult(out)
	os.Exit(exitCode(result, *strict))
}

// printResult prints data (a JSON document) piped through jq for pretty
// printing, falling back to a plain print if jq is unavailable or fails.
func printResult(data []byte) {
	cmd := exec.Command("jq", ".")
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Print(string(data))
	}
}
