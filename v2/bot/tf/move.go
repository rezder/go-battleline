package tf

import (
	"bytes"
	"encoding/binary"
	"github.com/rezder/go-battleline/v2/bot/prob"
	fa "github.com/rezder/go-battleline/v2/bot/prob/flag"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-error/log"
)

//Move finds the move that compares best to all the other moves.
func MoveHand(viewPos *game.ViewPos, tfCon *Con) (moveix int) {
	scoutMoveix := -1
	for ix, move := range viewPos.Moves {
		if move.IsScout() {
			scoutMoveix = ix
			break
		}
	}
	if scoutMoveix != -1 {
		moveix = scoutMoveix
	} else {
		if len(viewPos.Moves) == 1 {
			moveix = 0
		} else {
			tfAnas, _ := CalcTfAnas(viewPos, nil)
			bs, no := movesToBytes(tfAnas)
			log.Printf(log.Debug, "Sending %v moves", no)
			probs, err := tfCon.ReqProba(bs, no)
			if err != nil {
				log.PrintErr(err)
				log.Print(log.Min, "failed to make tenserflow move use prob move")
				moveix = prob.MoveHand(viewPos)
			} else {
				log.Printf(log.Debug, "Receiving %v probs", len(probs))
				m := make([][]float64, len(viewPos.Moves))
				for i, _ := range m {
					m[i] = make([]float64, len(viewPos.Moves))
				}
				probix := 0
				for i := 0; i < len(tfAnas); i++ {
					for j := 0; j < i; j++ {
						if tfAnas[i] != nil && tfAnas[j] != nil {
							m[i][j] = probs[probix]
							m[j][i] = 1 - probs[probix]
							probix++
						}
					}
				}
				log.Printf(log.Debug, "Probability matrix: %v", m)
				var max, sum float64
				for i, row := range m {
					for _, cell := range row {
						sum = sum + cell
					}
					if sum > max {
						max = sum
						moveix = i
					}
				}
			}
		}
	}
	log.Printf(log.Debug, "Moveix: %v Move: %v", moveix, viewPos.Moves[moveix])
	return moveix
}
func movesToBytes(tfAnas [][]*fa.TfAna) ([]byte, int) {

	var b []byte
	buf := bytes.NewBuffer(b)
	no := 0
	for i := 0; i < len(tfAnas); i++ {
		for j := 0; j < i; j++ {
			if tfAnas[i] != nil && tfAnas[j] != nil {
				no++
				for _, tfFlagAna := range tfAnas[i] {
					binary.Write(buf, binary.LittleEndian, tfFlagAna.MachineFloats())
				}
				for _, tfFlagAna := range tfAnas[j] {
					binary.Write(buf, binary.LittleEndian, tfFlagAna.MachineFloats())
				}
			}
		}
	}
	return buf.Bytes(), no
}
