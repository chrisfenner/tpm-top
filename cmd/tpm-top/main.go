package main

import (
	"fmt"
	"os"
	"time"

	"github.com/chrisfenner/tpm-top/pkg/opener"
	ui "github.com/gizak/termui/v3"
)

func main() {
	conf := opener.TcpConfig{
		Address: "127.0.0.1:2321",
	}

	if err := ui.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing termui: %v\n", err)
	}
	defer ui.Close()

	pcrView := NewPcrView()
	go func() {
		for true {
			width, height := ui.TerminalDimensions()
			pcrView.SetRect(0, 0, width, height)
			tpm, err := opener.OpenTcpTpm(&conf)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening TPM simulator: %v\n", err)
				return
			}
			pcrView.Refresh(tpm)
			tpm.Close()
			ui.Render(pcrView)
			time.Sleep(1 * time.Second)
		}
	}()

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}
