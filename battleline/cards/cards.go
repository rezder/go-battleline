// Package cards contain the cards of battleline.
package cards

import (
	"errors"
)

const (
	//NOTroop is the number of troop cards.
	NOTroop = 60
	//NOTac is the number of tactic cards.
	NOTac = 10

	//Card types.

	CTTroop  = 1
	CTMorale = 2
	CTEnv    = 3
	CTGuile  = 4

	//Troops colors.

	COLNone   = 0
	COLGreen  = 1
	COLRed    = 2
	COLPurpel = 3
	COLYellow = 4
	COLBlue   = 5
	COLOrange = 6

	//Tactic cards.

	TCAlexander = 70
	TCDarius    = 69
	TC8         = 68
	TC123       = 67
	TCFog       = 66
	TCMud       = 65
	TCScout     = 64
	TCRedeploy  = 63
	TCDeserter  = 62
	TCTraitor   = 61
)

var (
	//Cards contain all battleline cards.
	Cards [NOTac + NOTroop + 1]Card // zero index is nil
	//FWedge the wedge formation.
	FWedge = Formation{
		Name:    "Wedge",
		Describ: "3(4) troops connected and same color. Like 2,1,3 or 3,2,1.",
		Value:   5,
	}
	//FPhalanx the phalanx formation.
	FPhalanx = Formation{
		Name:    "Phalanx",
		Describ: "3(4) troops same value. Like 10,10,10",
		Value:   4,
	}
	//FBattalion the battalion Order formation.
	FBattalion = Formation{
		Name:    "Battalion Order",
		Describ: "3(4) troops same color",
		Value:   3,
	}
	//FSkirmish the skirmish line formation.
	FSkirmish = Formation{
		Name:    "Skirmish Line",
		Describ: "3(4) troops connected. Like 2,1,3 or 3,1,2",
		Value:   2,
	}
	//FHost the jost formation.
	FHost = Formation{
		Name:    "Host",
		Describ: "Any troops",
		Value:   1,
	}
)

func init() {
	names := [...]string{"Elephants",
		"Charriots",
		"Heavy Cavalry",
		"Light Cavalry",
		"Hypaspist",
		"Phalangist",
		"Hoplites",
		"Javelineers",
		"Peltasts",
		"Skirmishers",
	}

	for i := 1; i < 7; i++ {
		for j := 1; j < 11; j++ {
			troop := new(Troop)
			troop.name = names[10-j]
			troop.describ = names[10-j]
			troop.color = i
			troop.value = j
			Cards[(i-1)*10+j] = *troop //zero is nil
		}
	}

	Cards[TCAlexander] = Tactic{
		name:    "Leader Alexander",
		describ: "Any troop value. Only one leader can be played by a player. Value and color is defined before when the flag is resolved and not when played.",
		ttype:   CTMorale,
	}
	Cards[TCDarius] = Tactic{
		name:    "Leader Darius",
		describ: Cards[TCAlexander].Describ(),
		ttype:   CTMorale,
	}

	Cards[TC8] = Tactic{
		name:    "Companion Cavalry",
		describ: "Any color value 8. Color is defined when flag is resolved",
		ttype:   CTMorale,
	}
	Cards[TC123] = Tactic{
		name:    "Shield Bearers",
		describ: "Any color value 1, 2 or 3. Color is defined when flag is resolve",
		ttype:   CTMorale,
	}
	Cards[TCFog] = Tactic{
		name:    "Fog",
		describ: "Disables formation Flag is won by sum of values",
		ttype:   CTEnv,
	}

	Cards[TCMud] = Tactic{
		name:    "Mud",
		describ: "Formation is extended to 4 cards",
		ttype:   CTEnv,
	}
	Cards[TCScout] = Tactic{
		name:    "Scout",
		describ: "Draw 3 cards any decks and return 3 cards any decks player control the order",
		ttype:   CTGuile,
	}
	Cards[TCRedeploy] = Tactic{
		name:    "Redeploy",
		describ: "Move any of his troop or tactic card from any flag that is not claimed. Troop may be removed from game",
		ttype:   CTGuile,
	}
	Cards[TCDeserter] = Tactic{
		name:    "Deserter",
		describ: "Remove any opponent troop or tactic card from unclaimed flags",
		ttype:   CTGuile,
	}
	Cards[TCTraitor] = Tactic{
		name:    "Traitor",
		describ: "Take a troop from an opponents unclaimed flags and play it. You must have a slot to play the flag on",
		ttype:   CTGuile,
	}
}

//Card interface is the interface a battleline card.
type Card interface {
	Name() string
	Describ() string
	Type() int
}

//Troop is troop card.
type Troop struct {
	name    string
	describ string
	color   int
	value   int
}

//Name is the name of the troop card.
func (t Troop) Name() string {
	return t.name
}

//Describ is a short describtion of the troop card.
func (t Troop) Describ() string {
	return t.describ
}

//Type is CTTroop
func (t Troop) Type() int {
	return CTTroop
}

//Color is the color of the troop.
func (t Troop) Color() int {
	return t.color
}

//Value is the value of the troop 1-10.
func (t Troop) Value() int {
	return t.value
}

//Tactic is tactic card.
type Tactic struct {
	name    string
	describ string
	ttype   int
}

//Name is the name of the tactic card.
func (t Tactic) Name() string {
	return t.name
}

//Describ is short text describtion of the Tactic card.
func (t Tactic) Describ() string {
	return t.describ
}

//Type is CTMorale, CTEnv or CTGuile for a
//tactic card.
func (t Tactic) Type() int {
	return t.ttype
}

//Formation a flag formation.
type Formation struct {
	Name, Describ string
	Value         int
}

//DrCard returns the Card of the card index.
func DrCard(ix int) (c Card, err error) {
	if ix > 0 && ix < NOTroop+NOTac+1 {
		return Cards[ix], err
	}
	return nil, errors.New("Card do not exist")
}

//DrTroop returns the Troop card of the card index.
func DrTroop(ix int) (c *Troop, err error) {
	troop, ok := Cards[ix].(Troop)
	if ok {
		return &troop, err
	}
	return nil, errors.New("Troop do not exist")
}

//DrTactic returns the Tactic card of the card index.
func DrTactic(ix int) (c *Tactic, err error) {
	tactic, ok := Cards[ix].(Tactic)
	if ok {
		return &tactic, err
	}
	return nil, errors.New("Card do not exist")
}

//IsEnv checks if card index is a environment tactic card.
//Mud or fog.
func IsEnv(ix int) bool {
	return ix == TCMud || ix == TCFog
}

//IsMorale checks if a card index is morale tactic card.
//All jokers.
func IsMorale(ix int) bool {
	return ix == TCAlexander || ix == TCDarius || ix == TC123 || ix == TC8
}
func MoraleMaxValue(cardix int) int {
	switch cardix {
	case TCAlexander:
		fallthrough
	case TCDarius:
		return 10
	case TC123:
		return 3
	case TC8:
		return 8
	default:
		panic("Not a moral card")
	}

}

//IsTac checks if a card index is a tactic card.
func IsTac(ix int) bool {
	return ix > NOTroop
}

//IsTroop checks if a cardix is troop card.
func IsTroop(ix int) bool {
	return ix <= NOTroop
}
