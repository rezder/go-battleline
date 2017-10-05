// Package card contain the cards of battleline.
package card

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	//NOTroop is the number of troop cards.
	NOTroop = 60
	//NOTac is the number of tactic cards.
	NOTac = 10

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

	//BACKTac back of a tactic card.
	BACKTac = NOTac + NOTroop + 1
	//BACKTroop back of a troop card.
	BACKTroop = NOTac + NOTroop + 2
)

var (
	colors = [...]string{"None", "Green", "Red", "Purpel", "Yellow", "Blue", "Orange"}
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

//Card is the general card it include the back troop, back tactic
//and None. It just used to detect the different types of cards.
type Card uint8

func (m Card) Format(f fmt.State, c rune) {
	if c == 'v' && f.Flag('+') {
		switch {
		case m.IsTroop():
			fmt.Fprintf(f, "%v", Troop(m))
		case m.IsNone():
			io.WriteString(f, "None")
		case m.IsBack():
			fmt.Fprintf(f, "Back %v", Back(m))
		case m.IsEnv():
			fmt.Fprintf(f, "%v", Env(m))
		case m.IsMorale():
			fmt.Fprintf(f, "%v", Morale(m))
		case m.IsGuile():
			fmt.Fprintf(f, "%v", Guile(m))
		default:
			io.WriteString(f, "Illegal value not a Card!")
		}
	} else {
		fmt.Fprintf(f, "%v", uint8(m))
	}
}

//UnmarshalJSON unmarshalls json number to a
//card.
func (m *Card) UnmarshalJSON(b []byte) (err error) {
	var i int
	if err = json.Unmarshal(b, &i); err != nil {
		return err
	}
	*m = Card(i)
	return err
}

//MarshalJSON marshalls a card to json number to make
// a uint8 (byte) readable as a number.
func (m Card) MarshalJSON() (bytes []byte, err error) {
	return json.Marshal(int(m))
}

//IsUndefined return true if the card does not have a legal value.
// Warnning 0 the none value is included.
func (m Card) IsUndefined() bool {
	if m >= 0 && m < 73 {
		return true
	}
	return false
}

//IsTroop checks if a card move is troop card.
func (m Card) IsTroop() bool {
	if m > 0 && m <= NOTroop {
		return true
	}
	return false
}

//IsTac checks if a card move is a tactic card.
func (m Card) IsTac() bool {
	if NOTroop < m && m <= NOTac+NOTroop {
		return true
	}
	return false
}

//IsMorale checks if a card move is morale tactic card.
//All jokers.
func (m Card) IsMorale() bool {
	return m == TCAlexander || m == TCDarius || m == TC123 || m == TC8
}

//IsGuile checks if card move is guile card.
func (m Card) IsGuile() bool {
	return m == TCScout || m == TCDeserter || m == TCTraitor || m == TCRedeploy
}

//IsEnv checks if card indexmove is a environment tactic card.
//Mud or fog.
func (m Card) IsEnv() bool {
	return m == TCMud || m == TCFog
}

//IsBack checks if a card move is back of a card.
func (m Card) IsBack() bool {
	return m == BACKTac || m == BACKTroop
}

//IsNone checks if a card move is None
func (m Card) IsNone() bool {
	return m == 0
}

//Troop the troop card.
type Troop uint8

//UnmarshalJSON unmarshalls json number to
//a troop.
func (t *Troop) UnmarshalJSON(b []byte) (err error) {
	var i int
	if err = json.Unmarshal(b, &i); err != nil {
		return err
	}
	*t = Troop(i)
	return err
}

//MarshalJSON marshalls a troop to json number to make
// a uint8 (byte) readable as a number.
func (t Troop) MarshalJSON() (bytes []byte, err error) {
	return json.Marshal(int(t))
}

//Strenght is the strenght of the troop.
func (t Troop) Strenght() (s int) {
	s = int(t) % 10
	if s == 0 {
		s = 10
	}
	return s
}

//Color is the color of the troop.
func (t Troop) Color() int {
	return ((int(t) - 1) / 10) + 1
}

//TroopAppendSorted add troops in order of strenght.
func (t Troop) AppendStrSorted(troops []Troop) []Troop {
	no := len(troops)
	troops = append(troops, 0)
	for i, listTroop := range troops {
		if i == no {
			troops[i] = t
		} else {
			if !(t.Strenght() < listTroop.Strenght()) {
				copy(troops[i+1:], troops[i:])
				troops[i] = t
				break
			}
		}
	}
	return troops
}
func (t Troop) String() (txt string) {
	return fmt.Sprintf("%v %v", colors[t.Color()], t.Strenght())
}

//Morale the morale card.
type Morale uint8

//UnmarshalJSON unmarshalls json number to
//morale card.
func (m *Morale) UnmarshalJSON(b []byte) (err error) {
	var i int
	if err = json.Unmarshal(b, &i); err != nil {
		return err
	}
	*m = Morale(i)
	return err
}

//MarshalJSON marshalls a morale tactic card to json number to make
// a uint8 (byte) readable as a number.
func (m Morale) MarshalJSON() (bytes []byte, err error) {
	return json.Marshal(int(m))
}

//IsLeader checks if card is morale leader tactic card.
func (m Morale) IsLeader() bool {
	return m == TCAlexander || m == TCDarius
}

//HasStrenght is true if the morale card only
//have one strenght (currently only 8).
func (m Morale) HasStrenght() bool {
	return m == TC8
}

//MaxStrenght returns the maximum strenght a morale card can have.
func (m Morale) MaxStrenght() int {
	return m.Strenghts()[len(m.Strenghts())-1]
}

//ValidStrenght returns true if the morale card can
//cover the strenght.
func (m Morale) ValidStrenght(strenght int) bool {
	for _, s := range m.Strenghts() {
		if s == strenght {
			return true
		}
	}
	return false
}

//Strenghts returns the possible strenghts of the morale
//card.
func (m Morale) Strenghts() (sts []int) {
	switch m {
	case TCAlexander:
		fallthrough
	case TCDarius:
		sts = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	case TC123:
		sts = []int{1, 2, 3}
	case TC8:
		sts = []int{8}
	default:
		panic("Not a moral card")
	}
	return sts
}
func (m Morale) String() string {
	switch m {
	case TCAlexander:
		return "Alexander"
	case TCDarius:
		return "Darius"
	case TC123:
		return "123"
	case TC8:
		return "8"
	default:
		panic("Not a morale card.")
	}
}

//Env the enviroment card.
type Env uint8

//UnmarshalJSON unmarshalls json number to
//a tactic enviroment card.
func (e *Env) UnmarshalJSON(b []byte) (err error) {
	var i int
	if err = json.Unmarshal(b, &i); err != nil {
		return err
	}
	*e = Env(i)
	return err
}

//MarshalJSON marshalls a enviroment tactic card to json number to make
// a uint8 (byte) readable as a number.
func (e Env) MarshalJSON() (bytes []byte, err error) {
	return json.Marshal(int(e))
}

//IsFog returns true if enviroment card is fog.
func (e Env) IsFog() bool {
	return e == TCFog
}

//IsMud returns true if enviroment card is mud.
func (e Env) IsMud() bool {
	return e == TCMud
}
func (e Env) String() string {
	switch e {
	case TCMud:
		return "Mud"
	case TCDeserter:
		return "Fog"
	default:
		panic("Not a envirioment card.")
	}
}

//The Guile card.
type Guile uint8

func (e Guile) IsTraitor() bool {
	return e == TCTraitor
}
func (e Guile) IsDeserter() bool {
	return e == TCDeserter
}
func (e Guile) IsScout() bool {
	return e == TCScout
}
func (e Guile) IsRedeploy() bool {
	return e == TCRedeploy
}
func (e Guile) String() string {
	switch e {
	case TCTraitor:
		return "Traitor"
	case TCDeserter:
		return "Deserter"
	case TCRedeploy:
		return "Redeploy"
	case TCScout:
		return "Scout"
	default:
		panic("Not a Guile card.")
	}
}

//Back is the back of a card used in moves when the card is not known yet or
// in position view to hide the full information of scout returned cards.
type Back uint8

// IsTroop returns true if back represent a troop card.
func (b Back) IsTroop() bool {
	return b == BACKTroop
}

// IsTac returns true if back represent a tactic card.
func (b Back) IsTac() bool {
	return b == BACKTac
}
func (b Back) String() string {
	switch b {
	case BACKTac:
		return "Tactic"
	case BACKTroop:
		return "Troop"
	default:
		panic("Not a back card.")
	}
}

//Formation a flag formation.
type Formation struct {
	Name, Describ string
	Value         int
}

//Cards sorted cards.
type Cards struct {
	Troops  []Troop
	Morales []Morale
	Guiles  []Guile
	Envs    []Env
}

//Copy makes a copy of the silces.
func (c *Cards) Copy() (copyCards *Cards) {
	copyCards = new(Cards)

	copyCards.Troops = make([]Troop, len(c.Troops), cap(c.Troops))
	copy(copyCards.Troops, c.Troops)

	copyCards.Morales = make([]Morale, len(c.Morales), cap(c.Morales))
	copy(copyCards.Morales, c.Morales)

	copyCards.Guiles = make([]Guile, len(c.Guiles), cap(c.Guiles))
	copy(copyCards.Guiles, c.Guiles)

	copyCards.Envs = make([]Env, len(c.Envs), cap(c.Envs))
	copy(copyCards.Envs, c.Envs)

	return copyCards
}

//NoTacs the number of tactic cards.
func (c *Cards) NoTacs() int {
	return len(c.Morales) + len(c.Envs) + len(c.Guiles)
}

//No the number of cards.
func (c *Cards) No() int {
	return c.NoTacs() + len(c.Troops)
}

//Tacs mashes the tactic cards together
func (c *Cards) Tacs() (tacs []Card) {
	tacs = make([]Card, 0, c.NoTacs())
	for _, morale := range c.Morales {
		tacs = append(tacs, Card(morale))
	}
	for _, guile := range c.Guiles {
		tacs = append(tacs, Card(guile))
	}
	for _, env := range c.Envs {
		tacs = append(tacs, Card(env))
	}
	return tacs
}
func (c *Cards) Contain(srcCard Card) bool {
	switch {
	case srcCard.IsEnv():
		srcEnv := Env(srcCard)
		for _, env := range c.Envs {
			if env == srcEnv {
				return true
			}
		}
	case srcCard.IsTroop():
		srcTroop := Troop(srcCard)
		for _, troop := range c.Troops {
			if troop == srcTroop {
				return true
			}
		}
	case srcCard.IsMorale():
		srcMorale := Morale(srcCard)
		for _, morale := range c.Morales {
			if morale == srcMorale {
				return true
			}
		}
	case srcCard.IsGuile():
		srcGuile := Guile(srcCard)
		for _, guile := range c.Guiles {
			if guile == srcGuile {
				return true
			}
		}
	}
	return false
}
