package pcrs

import (
	"github.com/google/go-tpm/tpm2"
	"io"
)

// GetAlgorithms gets all the active PCR algorithms on the TPM.
func GetAlgorithms(tpm io.ReadWriter) ([]tpm2.Algorithm, error) {
	result := make([]tpm2.Algorithm, 0)
	// If a TPM has more than 8 active PCR banks, we will just have to miss out.
	// TODO: this more elegantly.
	sels, _, err := tpm2.GetCapability(tpm, tpm2.CapabilityPCRs, 8, 0)
	if err != nil {
		return nil, err
	}
	for _, sel := range sels {
		pcr, ok := sel.(tpm2.PCRSelection)
		if !ok {
			return nil, err
		}
		if len(pcr.PCRs) > 0 {
			result = append(result, pcr.Hash)
		}
	}
	return result, nil
}
