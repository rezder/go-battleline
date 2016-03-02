// The cards of the game

package battleline

const (
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

	TC_Alexander = 0
	TC_Darius    = 1
	TC_8         = 2
	TC_123       = 3
	TC_Fog       = 4
	TC_Mud       = 5
	TC_Scout     = 6
	TC_Redeploy  = 7
	TC_Deserter  = 8
	TC_Traitor   = 9
)

var (
	troops  [60]Troop
	tactics [10]Tactic
)

func init() {
	names := [...]string{"Elephants",
		"Charriots",
		"Heavy Cavalry",
		"Light Cavalry",
		"Hypaspist",
		"Phalangist",
		"Hoplites",
		"Javalineers",
		"Peltasts",
		"Skirmishers",
	}

	for i := 1; i < 7; i++ {
		for j := 1; i < 11; j++ {
			troop := troops[(i-1)*10+j-1]
			troop.name = names[10-j]
			troop.describ = names[10-j]
			troop.color = i
			troop.value = j
		}
	}

	tactic := tactics[0]
	tactic.name = "Leader Alexander"
	tactic.describ = "Any troop value. Only one leader can be played by a player. Value and color is defined before when the flag is resolved and not when played."
	tactic.ttype = TC_Alexander

	tactic = tactics[1]
	tactic.ttype = TC_Darius
	tactic.name = "Leader Darius"
	tactic.describ = tactics[0].describ

	tactic = tactics[2]
	tactic.name = "Companion Cavalry"
	tactic.describ = "Any color value 8. Color is defined when flag is resolved"
	tactic.ttype = TC_8

	tactic = tactics[3]
	tactic.name = "Shield Bearers"
	tactic.describ = "Any color value 1, 2 or 3. Color is defined when flag is resolve"
	tactic.ttype = TC_123

	tactic = tactics[4]
	tactic.name = "Fog"
	tactic.describ = "Disables formation Flag is won by sum of values"
	tactic.ttype = TC_Fog

	tactic = tactics[5]
	tactic.name = "Mud"
	tactic.describ = "Formation is extended to 4 cards"
	tactic.ttype = TC_Mud

	tactic = tactics[6]
	tactic.name = "Scout"
	tactic.describ = "Draw 3 cards any decks and return 3 cards any decks player control the order"
	tactic.ttype = TC_Scout

	tactic = tactics[7]
	tactic.name = "Redeploy"
	tactic.describ = "Move any of his troop or tactic card from any flag that is not claimed. Troop may be removed from game"
	tactic.ttype = TC_Redeploy

	tactic = tactics[8]
	tactic.name = "Deserter"
	tactic.describ = "Remove any opponent troop or tactic card from unclaimed flags"
	tactic.ttype = TC_Deserter

	tactic = tactics[9]
	tactic.name = "Traitor"
	tactic.describ = "Take a troop from an opponents unclaimed flags and play it. You must have a slot to play the flag on"
	tactic.ttype = TC_Traitor

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
	switch t.ttype {
	case TC_Alexander, TC_Darius, TC_8, TC_123:
		return T_Morale
	case TC_Fog, TC_Mud:
		return T_Env
	case TC_Scout, TC_Redeploy, TC_Deserter, TC_Traitor:
		return T_Guile
	}
}
func (t Tactic) TacticType() int {
	return t.ttype
}
