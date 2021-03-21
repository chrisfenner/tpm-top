package main

import (
	"fmt"
	"github.com/chrisfenner/tpm-top/internal/opener"
	"github.com/google/go-tpm/tpm2"
	"io"
	"os"
)

type toolFunc func(io.ReadWriter, []string) int

var funcMap = map[string]toolFunc{
	"startup": startup,
}

func startup(tpm io.ReadWriter, args []string) int {
	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "'startup' command expects 0 arguments")
		return 1
	}
	err := tpm2.Startup(tpm, tpm2.StartupClear)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calling TPM2_Startup: %v\n", err)
		return 1
	}
	return 0
}

func usage() {
	fmt.Printf("tpm-tool usage: tpm-tool (function) [(arguments)]\n")
	fmt.Printf("Supported functions:\n")
	for name, _ := range funcMap {
		fmt.Printf("  %s\n", name)
	}
}

func mainWithExitCode() int {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Please specify a command.\n")
		usage()
		return 1
	}
	cmd := os.Args[1]
	fun, ok := funcMap[cmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unsupported command '%s'\n", cmd)
		usage()
		return 1
	}

	conf := opener.TcpConfig{
		Address: "127.0.0.1:2321",
	}
	tpm, err := opener.OpenTcpTpm(&conf)
	if err != nil {
		fmt.Printf("Error opening TPM simulator: %v\n", err)
		return 1
	}
	defer tpm.Close()

	return fun(tpm, os.Args[2:])
}

func main() {
	os.Exit(mainWithExitCode())
}
