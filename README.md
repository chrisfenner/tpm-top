# tpm-top

tpm-top is a Go-based terminal utility like `top` for TPM. It displays
information about the TPM to which it is connected.

## Views
### PCRs
In this view, tpm-top displays all the PCR values that can fit into the window.
It live-updates every 1 second, reflecting the current state of all the PCRs.

NOTE: The Microsoft TPM Simulator comes by default with SHA1 and SHA2-256 banks
enable. Use `tpm-tool pcr-banks` (below) and reboot the simulator to pick just
one PCR bank.

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

## `tpm-tool` Commands
`tpm-tool` supports the following commands:

* `startup`
  * Starts up the TPM.
* `shutdown`
  * Shuts down the TPM.
* `pcr-banks <alg1> <alg2>...`
  * Enables the PCR banks for the given algorithm(s).
  * NOTE: The change will not take effect until you power cycle the TPM. You can do this with:
    * `tpm-tool shutdown`
    * `sim-start`
    * `tpm-tool startup`
* `extend <index> <file>`
  * Extends the contents of `<file>` into PCR `<index>` in all active PCR banks.
  * `<file>` must be 1KB or smaller.

## Starting the simulator
tpm-top currently only connects to a running TCP simulator, even if there is a
perfectly good local TPM on the system.
* Clone [the Microsoft reference implementation](https://github.com/microsoft/ms-tpm-20-ref).
* Build the simulator using the instructions from that repository.
* Start the simulator from the command-line.
* Use `sim-start` and `tpm-tool startup` to power-on the simulated TPM.

