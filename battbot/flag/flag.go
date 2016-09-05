package flag

import (
	"github.com/rezder/go-battleline/battleline/cards"
	"github.com/rezder/go-battleline/battserver/tables"
	slice "github.com/rezder/go-slice/int"
)

const (
	C_Opp   = -1
	C_Play  = 1
	COLNone = 0
)

type Flag struct {
	OppTroops  []int
	OppEnvs    []int
	PlayEnvs   []int
	PlayTroops []int
	Claimed    int
}

func New() (flag *Flag) {
	flag = new(Flag)
	flag.OppEnvs = make([]int, 0, 2)
	flag.PlayEnvs = make([]int, 0, 2)
	flag.OppTroops = make([]int, 0, 4)
	flag.PlayTroops = make([]int, 0, 4)
	return flag
}
func TransferTableFlag(tableFlag *tables.Flag) (flag *Flag) {
	flag = New()
	if tableFlag.OppFlag {
		flag.Claimed = C_Opp
	}
	if tableFlag.PlayFlag {
		flag.Claimed = C_Play
	}
	for _, v := range tableFlag.PlayTroops {
		flag.PlayTroops = slice.AddSorted(flag.PlayTroops, v, true)
	}
	for _, v := range tableFlag.OppTroops {
		flag.OppTroops = slice.AddSorted(flag.OppTroops, v, true)
	}
	for _, v := range tableFlag.OppEnvs {
		flag.OppEnvs = append(flag.OppEnvs, v)
	}
	for _, v := range tableFlag.PlayEnvs {
		flag.PlayEnvs = append(flag.PlayEnvs, v)
	}
	return flag
}
func (flag *Flag) PlayAddCardix(cardix int) {
	flag.OppTroops, flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs = addCard(flag.OppTroops,
		flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs, cardix, true)
}
func (flag *Flag) OppAddCardix(cardix int) {
	flag.OppTroops, flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs = addCard(flag.OppTroops,
		flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs, cardix, false)
}
func addCard(oppTroops, oppEnvs, playTroops, playEnvs []int, cardix int, player bool) ([]int, []int, []int, []int) {
	if cards.IsEnv(cardix) {
		if player {
			playEnvs = append(playEnvs, cardix)
		} else {
			oppEnvs = append(oppEnvs, cardix)
		}
	} else { //Troops
		if player {
			playTroops = slice.AddSorted(playTroops, cardix, true)
		} else {
			oppTroops = slice.AddSorted(oppTroops, cardix, true)
		}
	}
	return oppTroops, oppEnvs, playTroops, playEnvs
}
func (flag *Flag) OppRemoveCardix(cardix int) (found bool) {
	flag.OppTroops, flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs, found = removeCard(flag.OppTroops,
		flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs, cardix, false)
	return found
}
func (flag *Flag) PlayRemoveCardix(cardix int) (found bool) {
	flag.OppTroops, flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs, found = removeCard(flag.OppTroops,
		flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs, cardix, true)
	return found
}
func removeCard(oppTroops, oppEnvs, playTroops, playEnvs []int, cardix int, player bool) ([]int,
	[]int, []int, []int, bool) {
	updated := false
	if cards.IsEnv(cardix) {
		if player {
			playEnvs, updated = slice.RemoveV(playEnvs, cardix)
		} else {
			oppEnvs, updated = slice.RemoveV(oppEnvs, cardix)
		}
	} else { //Troops
		if player {
			playTroops, updated = slice.RemoveV(playTroops, cardix)
		} else {
			oppTroops, updated = slice.RemoveV(oppTroops, cardix)
		}
	}

	return oppTroops, oppEnvs, playTroops, playEnvs, updated
}
