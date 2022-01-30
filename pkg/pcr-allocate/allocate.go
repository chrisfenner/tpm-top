package pcrAllocate

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/chrisfenner/tpm-top/pkg/pcrs"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

const (
	cmdPcrAllocate tpmutil.Command = 0x12b
	rhPlatform     tpmutil.Handle  = 0x4000000c
)

// Removes items that are present in both lists of algorithms.
// Useful for removing redundant PCR bank selections below.
func dedup(l1, l2 []tpm2.Algorithm) ([]tpm2.Algorithm, []tpm2.Algorithm) {
	counts := make(map[tpm2.Algorithm]int)
	for _, alg := range l1 {
		counts[alg]++
	}
	for _, alg := range l2 {
		counts[alg]++
	}
	r1 := make([]tpm2.Algorithm, 0)
	for _, alg := range l1 {
		if counts[alg] != 2 {
			r1 = append(r1, alg)
		}
	}
	r2 := make([]tpm2.Algorithm, 0)
	for _, alg := range l2 {
		if counts[alg] != 2 {
			r2 = append(r2, alg)
		}
	}
	return r1, r2
}

type pcrAllocateResponse struct {
	AllocationSuccess uint8
	MaxPcr            uint32
	SizeNeeded        uint32
	SizeAvailable     uint32
}

// PcrAllocate allocates 24 PCRs for each of the given algorithms.
// Requires physical presence.
func PcrAllocate(tpm io.ReadWriter, algs []tpm2.Algorithm) error {
	// TPM2_PCR_ALLOCATE doesn't do anything to current PCR banks
	// that are not explicitly mentioned in the command.
	activeAlgs, err := pcrs.GetAlgorithms(tpm)
	if err != nil {
		return err
	}
	activeAlgs, algs = dedup(activeAlgs, algs)
	auth, err := encodeAuth()
	if err != nil {
		return err
	}
	parms, err := encodePcrSelections(activeAlgs, algs)
	if err != nil {
		return err
	}
	rsp, code, err := tpmutil.RunCommand(tpm, tpm2.TagSessions, cmdPcrAllocate, rhPlatform, tpmutil.RawBytes(auth), tpmutil.RawBytes(parms))
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("TPM2_PCR_ALLOCATE returned 0x%x", code)
	}
	// Strangely, we see an extra 4 bytes at the front of the response
	// and an extra 5 at the end.
	// The 4 seem to match the exact size of the expected response structure (0x0000000d).
	// The 5 don't make sense.
	// Otherwise, the data we see is exactly what we expect from the simulator.
	// TODO: make sense of this.
	reader := bytes.NewReader(rsp)
	var parsedSize uint32
	if err := binary.Read(reader, binary.BigEndian, &parsedSize); err != nil {
		return err
	}
	var parsedRsp pcrAllocateResponse
	if err := binary.Read(reader, binary.BigEndian, &parsedRsp); err != nil {
		return err
	}
	if parsedRsp.AllocationSuccess != 1 {
		return fmt.Errorf("TPM2_PCR_ALLOCATE returned RC_SUCCESS, but the following information:\n%+v\n", parsedRsp)
	}
	return nil
}

// encodeAuth encodes a TPM 2.0 command auth area with a single password session using the empty Platform auth value.
func encodeAuth() ([]byte, error) {
	auth := tpm2.AuthCommand{
		Session:    tpm2.HandlePasswordSession,
		Attributes: tpm2.SessionAttributes(0),
		Auth:       tpm2.EmptyAuth,
	}
	authBuf, err := tpmutil.Pack(auth)
	if err != nil {
		return nil, err
	}
	size, err := tpmutil.Pack(uint32(len(authBuf)))
	if err != nil {
		return nil, err
	}
	authBuf = append(size, authBuf...)
	return authBuf, nil
}

// encodePcrSelections encodes a TPML_PCR_SELECTION of 24 PCRs in each of the given algorithms.
func encodePcrSelections(remove, add []tpm2.Algorithm) ([]byte, error) {
	selection := &bytes.Buffer{}
	// TPML_PCR_SELECTION.count
	if err := binary.Write(selection, binary.BigEndian, uint32(len(add)+len(remove))); err != nil {
		return nil, err
	}
	// TPML_PCR_SELECTION.pcrSelections
	for _, alg := range remove {
		// TPMS_PCR_SELECTION.hash
		if err := binary.Write(selection, binary.BigEndian, alg); err != nil {
			return nil, err
		}
		// TPMS_PCR_SELECTION.sizeOfSelect
		if err := binary.Write(selection, binary.BigEndian, uint8(3)); err != nil {
			return nil, err
		}
		// TPMS_PCR_SELECTION.pcrSelect
		for i := 0; i < 3; i++ {
			if err := binary.Write(selection, binary.BigEndian, uint8(0x00)); err != nil {
				return nil, err
			}
		}
	}
	// TPML_PCR_SELECTION.pcrSelections
	for _, alg := range add {
		// TPMS_PCR_SELECTION.hash
		if err := binary.Write(selection, binary.BigEndian, alg); err != nil {
			return nil, err
		}
		// TPMS_PCR_SELECTION.sizeOfSelect
		if err := binary.Write(selection, binary.BigEndian, uint8(3)); err != nil {
			return nil, err
		}
		// TPMS_PCR_SELECTION.pcrSelect
		for i := 0; i < 3; i++ {
			if err := binary.Write(selection, binary.BigEndian, uint8(0xff)); err != nil {
				return nil, err
			}
		}
	}
	return selection.Bytes(), nil
}
