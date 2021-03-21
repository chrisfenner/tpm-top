package main

import (
	"encoding/hex"
	"fmt"
	"github.com/chrisfenner/tpm-top/internal/opener"
	"github.com/google/go-tpm/tpm2"
)

func main() {
	conf := opener.TcpConfig{
		Address: "127.0.0.1:2321",
	}
	tpm, err := opener.OpenTcpTpm(&conf)
	if err != nil {
		fmt.Printf("Error opening TPM simulator: %v\n", err)
		return
	}
	defer tpm.Close()
	mfr, err := tpm2.GetManufacturer(tpm)
	if err != nil {
		fmt.Printf("Error calling GetManufacturer: %v\n", err)
		return
	}
	fmt.Printf("Manufacturer string: %s\n", hex.EncodeToString(mfr))
}
