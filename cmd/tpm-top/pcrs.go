package main

import (
	"encoding/hex"
	"fmt"
	"github.com/chrisfenner/tpm-top/internal/pcrs"
	ui "github.com/gizak/termui/v3"
	"github.com/google/go-tpm/tpm2"
	"image"
	"io"
)

var (
	pcrAlgStyle = ui.Style{
		Fg:       226,
		Bg:       234,
		Modifier: ui.ModifierBold | ui.ModifierUnderline,
	}
	pcrDataStyle = ui.Style{
		Fg:       15,
		Bg:       ui.ColorClear,
		Modifier: ui.ModifierClear,
	}
	pcrIndexStyle = ui.Style{
		Fg:       14,
		Bg:       ui.ColorClear,
		Modifier: ui.ModifierBold,
	}
	paddingStyle = ui.Style{
		Fg:       ui.ColorClear,
		Bg:       ui.ColorClear,
		Modifier: ui.ModifierClear,
	}
)

// PcrView is a PCR-showing widget that can be used with termui.
type PcrView struct {
	ui.Block
	pcrs []pcrData
}

// NewPcrView creates a new PcrView.
func NewPcrView() *PcrView {
	result := &PcrView{
		Block: *ui.NewBlock(),
	}
	result.Block.Title = "PCRs"
	return result
}

// Refresh refreshes the view with new data from the TPM.
func (p *PcrView) Refresh(tpm io.ReadWriter) {
	pcrBanks := make([]pcrData, 0)
	// Find out which PCRs are implemented.
	algs, err := pcrs.GetAlgorithms(tpm)
	if err != nil {
		panic(err)
	}
	// Read each PCR bank.
	for _, alg := range algs {
		pcrBanks = append(pcrBanks, pcrBank(tpm, alg))
	}
	p.pcrs = pcrBanks
}

// pcrBank reads all the PCRs in the selected bank.
func pcrBank(tpm io.ReadWriter, alg tpm2.Algorithm) pcrData {
	pcrs := make([][]byte, 24)
	// Read PCRs 8 at a time, the max supported by the TPM.
	for i := 0; i < 24; i += 8 {
		sel := tpm2.PCRSelection{
			Hash: alg,
			PCRs: []int{i, i + 1, i + 2, i + 3, i + 4, i + 5, i + 6, i + 7},
		}
		hashes, err := tpm2.ReadPCRs(tpm, sel)
		if err != nil {
			panic(err)
		}
		for idx, hash := range hashes {
			pcrs[idx] = hash
		}
	}
	return pcrData{
		alg:    alg,
		hashes: pcrs,
	}
}

// cellsFromPcr formats PCR data as an array of arrays of cells, ready to be drawn.
func cellsFromPcr(pcr *pcrData, w int) [][]ui.Cell {
	// Pretty-print the algorithm name.
	result := make([][]ui.Cell, 0)
	title := algName(pcr.alg)
	result = append(result, ui.RunesToStyledCells(title, pcrAlgStyle))

	for i, hash := range pcr.hashes {
		// For each PCR value, pretty-print the index and the data.
		// For some algorithms, the PCR text will need to span multiple lines.
		// In that case, don't print the index again, but leave space.
		idxWidth, paddingWidth, dataWidth := decidePcrWidths(w, len(hash))
		index := indexString(i, idxWidth)
		lines := pcrStrings(hash, dataWidth)
		for _, line := range lines {
			resultLine := ui.RunesToStyledCells(index, pcrIndexStyle)
			resultLine = append(resultLine, ui.RunesToStyledCells(make([]rune, paddingWidth), paddingStyle)...)
			resultLine = append(resultLine, ui.RunesToStyledCells(line, pcrDataStyle)...)
			result = append(result, resultLine)
			// Clear the index area for multi-line PCRs
			index = make([]rune, len(index))
		}
	}

	return result
}

// indexString formats the given index for pretty-printing, depending on the width.
func indexString(index, width int) []rune {
	if width >= 8 {
		return []rune(fmt.Sprintf("PCR[%02d]:", index))
	}
	return []rune(fmt.Sprintf("%02d:", index))
}

// pcrStrings formats the given hash for pretty-printing, depending on the width.
func pcrStrings(hash []byte, width int) [][]rune {
	result := make([][]rune, 0)
	hex := []rune(hex.EncodeToString(hash))
	for i := 0; i < len(hex); i += width {
		result = append(result, hex[i:i+width])
	}
	return result
}

// Draw implements the termui Drawable interface.
func (p *PcrView) Draw(buf *ui.Buffer) {
	p.Block.Draw(buf)
	y := p.Block.Inner.Min.Y - 1 // Inner.Min.Y inexplicably leaves an empty line.
	// For each PCR bank we have data for,
	for _, pcr := range p.pcrs {
		// Grab the rectangle of characters we want to print for this PCR bank.
		cells := cellsFromPcr(&pcr, p.Block.Inner.Dx())
		// For each row in the rectangle of cells,
		for _, row := range cells {
			// For each character in the row, write it to the buffer.
			for x, cell := range row {
				buf.SetCell(cell, image.Pt(x, y).Add(p.Block.Inner.Min))
			}
			y += 1
			// Bail out if we run out of vertical space.
			if y >= p.Block.Inner.Dy() {
				return
			}
		}
	}
}

// pcrData represents a PCR bank.
type pcrData struct {
	alg    tpm2.Algorithm
	hashes [][]byte
}

// decidePcrWidths decides the spacing to give to the PCR index, padding, and PCR data.
func decidePcrWidths(width int, dataLength int) (indexWidth, paddingWidth, dataWidth int) {
	// Roomy 1-line:
	// PCR[00]:  deadbeefdeadbeefdeadbeefdeadbeef
	// 8, 2, (2 * dataLength)
	if width >= (8 + 2 + (2 * dataLength)) {
		return 8, 2, 2 * dataLength
	}
	// Compact 1-line:
	// 00: deadbeefdeadbeefdeadbeefdeadbeef
	// 3, 1, (2 * dataLength)
	if width >= (3 + 1 + (2 * dataLength)) {
		return 3, 1, 2 * dataLength
	}
	// Roomy 2-lines:
	// PCR[00]:  deadbeefdeadbeef
	//           deadbeefdeadbeef
	// 3, 1, dataLength
	if width >= (8 + 2 + dataLength) {
		return 8, 2, dataLength
	}
	// Compact N-lines:
	// 00: deadbeef
	//     deadbeef
	//     deadbeef
	//     ...
	// 3, 1, leftover
	if width > (3 + 1) {
		return 3, 1, width - (3 + 1)
	}
	// Pathological case. Not going to look good at all.
	return 0, 0, width
}

// algName prints out the name of the PCR algorithm if it's known.
func algName(alg tpm2.Algorithm) []rune {
	var niceName string
	switch alg {
	case tpm2.AlgSHA1:
		niceName = "SHA1"
		break
	case tpm2.AlgSHA256:
		niceName = "SHA2-256"
		break
	case tpm2.AlgSHA384:
		niceName = "SHA2-384"
		break
	default:
		niceName = "UNKNOWN"
		break
	}
	return []rune(fmt.Sprintf("%s (0x%04x)", niceName, alg))
}
