// The cards of the game
package cards

import (
	"errors"
)

const (
	T  = 60
	TC = 10

	T_Troop  = 1
	T_Morale = 2
	T_Env    = 3
	T_Guile  = 4

	C_Green  = 1
	C_Red    = 2
	C_Purpel = 3
	C_Yellow = 4
	C_Blue   = 5
	C_Orange = 6

	TC_Alexander = 70
	TC_Darius    = 69
	TC_8         = 68
	TC_123       = 67
	TC_Fog       = 66
	TC_Mud       = 65
	TC_Scout     = 64
	TC_Redeploy  = 63
	TC_Deserter  = 62
	TC_Traitor   = 61
)

var (
	Cards   [TC + T + 1]Card // zero index is nil
	F_Wedge = Formation{
		Name:    "Wedge",
		Describ: "3(4) troops connected and same color. Like 2,1,3 or 3,2,1.",
		Value:   5,
	}
	F_Phalanx = Formation{
		Name:    "Phalanx",
		Describ: "3(4) troops same value. Like 10,10,10",
		Value:   4,
	}
	F_BattalionOrder = Formation{
		Name:    "Battalion Order",
		Describ: "3(4) troops same color",
		Value:   3,
	}
	F_SkirmishLine = Formation{
		Name:    "Skirmish Line",
		Describ: "3(4) troops connected. Like 2,1,3 or 3,1,2",
		Value:   2,
	}
	F_Host = Formation{
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
			var troop *Troop = new(Troop)
			troop.name = names[10-j]
			troop.describ = names[10-j]
			troop.color = i
			troop.value = j
			Cards[(i-1)*10+j] = *troop //zero is nil
		}
	}

	Cards[TC_Alexander] = Tactic{
		name:    "Leader Alexander",
		describ: "Any troop value. Only one leader can be played by a player. Value and color is defined before when the flag is resolved and not when played.",
		ttype:   T_Morale,
	}
	Cards[TC_Darius] = Tactic{
		name:    "Leader Darius",
		describ: Cards[TC_Alexander].Describ(),
		ttype:   T_Morale,
	}

	Cards[TC_8] = Tactic{
		name:    "Companion Cavalry",
		describ: "Any color value 8. Color is defined when flag is resolved",
		ttype:   T_Morale,
	}
	Cards[TC_123] = Tactic{
		name:    "Shield Bearers",
		describ: "Any color value 1, 2 or 3. Color is defined when flag is resolve",
		ttype:   T_Morale,
	}
	Cards[TC_Fog] = Tactic{
		name:    "Fog",
		describ: "Disables formation Flag is won by sum of values",
		ttype:   T_Env,
	}

	Cards[TC_Mud] = Tactic{
		name:    "Mud",
		describ: "Formation is extended to 4 cards",
		ttype:   T_Env,
	}
	Cards[TC_Scout] = Tactic{
		name:    "Scout",
		describ: "Draw 3 cards any decks and return 3 cards any decks player control the order",
		ttype:   T_Guile,
	}
	Cards[TC_Redeploy] = Tactic{
		name:    "Redeploy",
		describ: "Move any of his troop or tactic card from any flag that is not claimed. Troop may be removed from game",
		ttype:   T_Guile,
	}
	Cards[TC_Deserter] = Tactic{
		name:    "Deserter",
		describ: "Remove any opponent troop or tactic card from unclaimed flags",
		ttype:   T_Guile,
	}
	Cards[TC_Traitor] = Tactic{
		name:    "Traitor",
		describ: "Take a troop from an opponents unclaimed flags and play it. You must have a slot to play the flag on",
		ttype:   T_Guile,
	}
}

type Card interface {
	Name() string
	Describ() string
	Type() int
}

type Troop struct {
	name    string
	describ string
	color   int
	value   int
}

func (t Troop) Name() string {
	return t.name
}

func (t Troop) Describ() string {
	return t.describ
}
func (t Troop) Type() int {
	return T_Troop
}
func (t Troop) Color() int {
	return t.color
}
func (t Troop) Value() int {
	return t.value
}

type Tactic struct {
	name    string
	describ string
	ttype   int
}

func (t Tactic) Name() string {
	return t.name
}
func (t Tactic) Describ() string {
	return t.describ
}
func (t Tactic) Type() int {
	return t.ttype
}

type Formation struct {
	Name, Describ string
	Value         int
}

func DrCard(ix int) (c Card, err error) {
	if ix > 0 && ix < T+TC+1 {
		return Cards[ix], err
	} else {
		return nil, errors.New("Card do not exist")
	}
}

func DrTroop(ix int) (c *Troop, err error) {
	troop, ok := Cards[ix].(Troop)
	if ok {
		return &troop, err
	} else {
		return nil, errors.New("Troop do not exist")
	}
}
func DrTactic(ix int) (c *Tactic, err error) {
	tactic, ok := Cards[ix].(Tactic)
	if ok {
		return &tactic, err
	} else {
		return nil, errors.New("Card do not exist")
	}
}
