package flag

import (
	"github.com/rezder/go-battleline/battleline/cards"
	"github.com/rezder/go-battleline/battserver/tables"
	slice "github.com/rezder/go-slice/int"
)

const (
	CLAIMPlay = 1
	CLAIMOpp  = -1
)

//Flag a battleline flag.
type Flag struct {
	OppTroops  []int
	OppEnvs    []int
	PlayEnvs   []int
	PlayTroops []int
	Claimed    int
}

//New create a flag.
func New() (flag *Flag) {
	flag = new(Flag)
	flag.OppEnvs = make([]int, 0, 2)
	flag.PlayEnvs = make([]int, 0, 2)
	flag.OppTroops = make([]int, 0, 4)
	flag.PlayTroops = make([]int, 0, 4)
	return flag
}

//Copy copy flag
func (flag *Flag) Copy() (c *Flag) {
	if flag == nil {
		c = nil
	} else {
		c = New()
		c.OppEnvs = append(c.OppEnvs, flag.OppEnvs...)
		c.PlayEnvs = append(c.PlayEnvs, flag.PlayEnvs...)
		c.OppTroops = append(c.OppTroops, flag.OppTroops...)
		c.PlayTroops = append(c.PlayTroops, flag.PlayTroops...)
		c.Claimed = flag.Claimed
	}
	return c
}

//TransferTableFlag transfers a table flag to a flag.
func TransferTableFlag(tableFlag *tables.Flag) (flag *Flag) {

	flag = New()
	if tableFlag.OppFlag {
		flag.Claimed = CLAIMOpp
	}
	if tableFlag.PlayFlag {
		flag.Claimed = CLAIMPlay
	}
	for _, v := range tableFlag.PlayTroops {
		flag.PlayTroops = slice.AddSorted(flag.PlayTroops, v, true)
	}
	for _, v := range tableFlag.OppTroops {
		flag.OppTroops = slice.AddSorted(flag.OppTroops, v, true)
	}
	flag.OppEnvs = append(flag.OppEnvs, tableFlag.OppEnvs...)
	flag.PlayEnvs = append(flag.PlayEnvs, tableFlag.PlayEnvs...)
	return flag
}

// PlayAddCardix adds a player card to the flag.
func (flag *Flag) PlayAddCardix(cardix int) {
	flag.OppTroops, flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs = addCard(flag.OppTroops,
		flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs, cardix, true)
}

//OppAddCardix adds a opponent card to the flag.
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

// OppRemoveCardix removes a opponent card.
func (flag *Flag) OppRemoveCardix(cardix int) (found bool) {
	flag.OppTroops, flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs, found = removeCard(flag.OppTroops,
		flag.OppEnvs, flag.PlayTroops, flag.PlayEnvs, cardix, false)
	return found
}

// PlayRemoveCardix removes a player card.
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

// IsFog returns true if a flag contain tactic card fog.
func (flag *Flag) IsFog() bool {
	return flag.isEnv(cards.TCFog)
}
func (flag *Flag) isEnv(env int) bool {
	contain := false
	for _, cardix := range flag.OppEnvs {
		if env == cardix {
			contain = true
			break
		}
	}
	if !contain {
		for _, cardix := range flag.PlayEnvs {
			if env == cardix {
				contain = true
				break
			}
		}
	}
	return contain
}

// IsMud returns if a flag contain tactic card mud.
func (flag *Flag) IsMud() bool {
	return flag.isEnv(cards.TCMud)
}
func (flag *Flag) FormationSize() (size int) {
	size = 3
	if flag.IsMud() {
		size = 4
	}
	return size
}
func (flag *Flag) IsClaimed() bool {
	return flag.Claimed != 0
}
