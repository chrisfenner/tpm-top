# tpm-top

tpm-top is a Go-based command line utility like `top` for TPM. It displays
information about the TPM to which it is connected.

## Views
No views are supported yet. tpm-tool prints out the manufacturer ID as a basic
connectivity check.

## Supported TPM types
* TCP simulator (like [the Microsoft reference TPM 2.0](https://github.com/microsoft/ms-tpm-20-ref))

To aid development and demonstration of tpm-top, some additional tools are
included in this repository:
* tpm-tool
  * A tool that can send a few sample commands to a running TPM.
* sim-start
  * A tool that can (re)start a running TCP simulator.

## Building
* tpm-top (like all other tools in this repo) is built using `go build`, e.g.,
from the root of the repository, run:
```
go build cmd/tpm-top
```

## Starting the simulator
tpm-top currently only connects to a running TCP simulator, even if there is a
perfectly good local TPM on the system.
* Clone [the Microsoft reference implementation](https://github.com/microsoft/ms-tpm-20-ref)).
* Build the simulator using the instructions from that repository.
* Start the simulator from the command-line.

