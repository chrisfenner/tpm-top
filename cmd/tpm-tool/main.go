package main

import (
	"fmt"
	"github.com/chrisfenner/tpm-top/internal/opener"
	"github.com/chrisfenner/tpm-top/internal/pcr-allocate"
	"github.com/chrisfenner/tpm-top/internal/rc"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
	"io"
	"os"
	"strconv"
	"strings"
)

type toolFunc func(io.ReadWriter, []string) int

var funcMap = map[string]toolFunc{
	"startup":   startup,
	"shutdown":  shutdown,
	"pcr-banks": pcrBanks,
	"extend":    extend,
}

type toolFuncNoTpm func([]string) int

var funcMapNoTpm = map[string]toolFuncNoTpm{
	"explain":   explain,
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

func shutdown(tpm io.ReadWriter, args []string) int {
	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "'shutdown' command expects 0 arguments")
		return 1
	}
	err := tpm2.Shutdown(tpm, tpm2.StartupClear)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calling TPM2_Shutdown: %v\n", err)
		return 1
	}
	return 0
}

func pcrBanks(tpm io.ReadWriter, args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "'pcr-banks' command expects at least one hash algorithm\n")
		return 1
	}
	algs := make([]tpm2.Algorithm, 0)
	for _, algStr := range args {
		algStr = strings.ToLower(algStr)
		var alg tpm2.Algorithm
		switch algStr {
		case "sha1":
			alg = tpm2.AlgSHA1
		case "sha256":
			fallthrough
		case "sha2-256":
			alg = tpm2.AlgSHA256
		case "sha384":
			fallthrough
		case "sha2-384":
			alg = tpm2.AlgSHA384
		case "sha512":
			fallthrough
		case "sha2-512":
			alg = tpm2.AlgSHA512
		case "sha3-256":
			alg = tpm2.AlgSHA3_256
		case "sha3-384":
			alg = tpm2.AlgSHA3_384
		case "sha3-512":
			alg = tpm2.AlgSHA3_512
		default:
			fmt.Fprintf(os.Stderr, "Unrecognized hash algorithm '%s'\n", algStr)
			return 1
		}
		algs = append(algs, alg)
	}
	if err := pcrAllocate.PcrAllocate(tpm, algs); err != nil {
		fmt.Fprintf(os.Stderr, "Error calling TPM2_PCR_ALLOCATE: %v\n", err)
		return 1
	}
	return 0
}

func extend(tpm io.ReadWriter, args []string) int {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "'extend' command expects 2 arguments: a file and an index")
		return 1
	}
	pcrIndex, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse PCR index: %v\n", err)
		return 1
	}
	if pcrIndex < 0 || pcrIndex > 23 {
		fmt.Fprintf(os.Stderr, "PCR index must be between 0 and 23.\n")
		return 1
	}
	eventFile := args[1]
	f, err := os.Open(eventFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open %s: %v\n", eventFile, err)
		return 1
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not stat %s: %v\n", eventFile, err)
		return 1
	}
	if info.Size() > 1024 {
		fmt.Fprintf(os.Stderr, "%s is too large. Please pass a file of size 1KB or smaller.\n", eventFile)
		return 1
	}
	contents := make([]byte, info.Size())
	if _, err := f.Read(contents); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", eventFile, err)
		return 1
	}
	if err := tpm2.PCREvent(tpm, tpmutil.Handle(pcrIndex), contents); err != nil {
		fmt.Fprintf(os.Stderr, "Error in TPM2_PCR_EVENT: %v\n", err)
		return 1
	}

	return 0
}

func explain(args []string) int {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "'explain' command expects 1 argument: an error code\n")
		return 1
	}
	code, err := strconv.ParseInt(args[0], 0, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse error code: %v\n", err)
		return 1
	}
	err = rc.MakeError(int(code))
	if err == nil {
		fmt.Printf("RC_SUCCESS! 😎\n")
	} else {
		fmt.Printf("%s\n", err)
	}
	return 0
}

func usage() {
	fmt.Printf("tpm-tool usage: tpm-tool (function) [(arguments)]\n")
	fmt.Printf("Supported functions:\n")
	for name, _ := range funcMap {
		fmt.Printf("  %s\n", name)
	}
	for name, _ := range funcMapNoTpm {
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
	funNoTpm, ok := funcMapNoTpm[cmd]
	if ok {
		return funNoTpm(os.Args[2:])
	}
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
