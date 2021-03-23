package main

import (
	"fmt"
	"github.com/chrisfenner/tpm-top/internal/platform"
)

func main() {
	conf := platform.TcpConfig{
		Address: "127.0.0.1:2322",
	}
	platform, err := platform.OpenTcpPlatform(&conf)
	if err != nil {
		fmt.Printf("Error connecting to TPM simulator: %v\n", err)
		return
	}
	defer func() {
		if err := platform.Close(); err != nil {
			fmt.Printf("Error closing platform: %v\n", err)
		}
	}()

	platform.NVOff()
	platform.PowerOff()
	platform.PowerOn()
	platform.NVOn()
}
