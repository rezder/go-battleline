//The battleline flag.
package flag

import (
	"errors"
	"fmt"
	"rezder.com/game/card/battleline/cards"
	Math "rezder.com/math/int"
	slice "rezder.com/slice/int"
	"sort"
)

const (
	vm123    = 3
	vmLeader = 10
)

//Player the structer for player details of the flag.
type Player struct {
	Won       bool
	Env       [2]int
	Troops    [4]int
	Formation *cards.Formation
	Strenght  int
}

func (player *Player) Equal(other *Player) (equal bool) {
	if other == nil && player == nil {
		equal = true
	} else if other != nil && player != nil {
		if player == other {
			equal = true
		} else if other.Won == player.Won && other.Strenght == player.Strenght {
			if other.Formation == player.Formation || other.Formation.Value == player.Formation.Value {
				if slice.Equal(other.Env[:], player.Env[:]) && slice.Equal(other.Troops[:], player.Troops[:]) {
					equal = true
				}
			}
		}
	}
	return equal
}

//Flag the structer of the flag.
type Flag struct {
	Players [2]*Player
}

func New() (flag *Flag) {
	flag = new(Flag)
	flag.Players[0] = new(Player)
	flag.Players[1] = new(Player)
	return flag
}
func (flag *Flag) Copy() (c *Flag) {
	c = new(Flag)
	c.Players[0] = &*flag.Players[0]
	c.Players[1] = &*flag.Players[1]
	return c
}

func (f *Flag) Equal(o *Flag) (equal bool) {
	if o == nil && f == nil {
		equal = true
	} else if o != nil && f != nil {
		if o == f {
			equal = true
		} else {
			equal = true
			for i, v := range o.Players {
				if !v.Equal(f.Players[i]) {
					equal = false
					break
				}
			}
		}
	}
	return equal
}

// Remove removes a card from the flag.
// mudix0 contains a card if removal of the mud card result in an excess card for player 0/1
func (flag *Flag) Remove(cardix int, playerix int) (mudix0 int, mudix1 int, err error) {
	player := flag.Players[playerix]        // updated
	opp := flag.Players[opponent(playerix)] // updated
	mudix1 = -1
	mudix0 = -1
	card, err := cards.DrCard(cardix)
	if err == nil {
		errMessage := fmt.Sprintf("Player %v do not have card %v", playerix, card.Name())
		var delix int
		switch tcard := card.(type) {
		case cards.Tactic:
			if tcard.Type() == cards.T_Env {
				delix, err = remove_clear(player.Env[:], cardix, errMessage)
				if delix != -1 {
					if cardix == cards.TC_Mud {
						if played(player.Troops[:]) == 4 {
							mudix0 = remove_Mud(player, opp)
							remove_clear(player.Troops[:], mudix0, errMessage)
						}
						if played(opp.Troops[:]) == 4 {
							mudix1 = remove_Mud(opp, player)
							remove_clear(opp.Troops[:], mudix1, errMessage)
						}
					}
					sort.Sort(sort.Reverse(sort.IntSlice(player.Env[:])))
					player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
					opp.Formation, opp.Strenght = eval(opp.Troops[:], opp.Env[:], player.Env[:])
				}
			} else if tcard.Type() == cards.T_Morale {
				delix, err = remove_clear(player.Troops[:], cardix, errMessage)
				if delix != -1 {
					player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
				}
			} else {
				err = errors.New(fmt.Sprintf("Illegal tactic card: %v", card.Name()))
			}
		case cards.Troop:
			delix, err = remove_clear(player.Troops[:], cardix, errMessage)
			if delix != -1 {
				player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
			}

		default:
			panic("No supprted type")
		}
	} else {
		panic("Card do not exist")
	}
	return mudix0, mudix1, err
}

// remove_Mud finds the excess card that gives the highest formation
// and strenght when it is removed.
func remove_Mud(player *Player, opp *Player) (mudix int) {
	var maxf *cards.Formation = &cards.F_Host
	var maxs int = 1
	var troops []int
	for i, v := range player.Troops {
		troops = make([]int, 4)
		copy(troops, player.Troops[:])
		troops[i] = 0
		formation, strength := eval(troops, player.Env[:], opp.Env[:])
		if formation.Value > maxf.Value {
			maxf = formation
			maxs = strength
			mudix = v
		} else if formation.Value == maxf.Value && strength > maxs {
			maxs = strength
			mudix = v
		}
	}
	return mudix
}

// remove_clear zero a card index.
// ix the index in list if the card was found if not -1.
func remove_clear(cards []int, cardix int, errM string) (ix int, err error) {
	ix = -1
	for i, v := range cards {
		if v == cardix {
			cards[i] = 0
			ix = i
			break
		}
	}
	if ix == -1 {
		err = errors.New(errM)
	}
	return ix, err
}

// opponent calculate the opponent player index.
func opponent(playerix int) int {
	if playerix == 0 {
		return 1
	} else {
		return 0
	}
}

//Set a card.
func (flag *Flag) Set(cardix int, playerix int) (err error) {
	player := flag.Players[playerix]        // updated
	opp := flag.Players[opponent(playerix)] // updated
	card, err := cards.DrCard(cardix)
	if err == nil {
		errMessage := fmt.Sprintf("Player %v do not have space for card %v", playerix, card.Name())
		var place int
		switch tcard := card.(type) {
		case cards.Tactic:
			if tcard.Type() == cards.T_Env {
				place, err = set_Card(cardix, player.Env[:], errMessage)
				if place != -1 {
					player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
					opp.Formation, opp.Strenght = eval(opp.Troops[:], opp.Env[:], player.Env[:])
				}
			} else if tcard.Type() == cards.T_Morale {
				place, err = set_Card(cardix, player.Troops[:], errMessage)
				if place != -1 {
					player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
				}
			} else {
				err = errors.New("Illegal tactic card: " + card.Name())
			}
		case cards.Troop:
			place, err = set_Card(cardix, player.Troops[:], errMessage)
			if place != -1 {
				player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
			}
		default:
			fmt.Printf("card type %v", tcard)
			panic("No supported type")
		}
	} else {
		panic("Card do not exist")
	}
	return err
}

// set_Card set a card in a list of cards.
// #v set a card at the first available spot.
func set_Card(cardix int, v []int, errM string) (place int, err error) {
	place = -1
	for i, c := range v {
		if c == 0 {
			v[i] = cardix
			place = i
			break
		}
	}
	if place == -1 {
		err = errors.New(errM)
	}
	return place, err
}

// played calculate the played cards of a players troops.
// DO not assume sort.
func played(troops []int) (no int) {
	no = 0
	for _, troop := range troops {
		if troop != 0 {
			no++
		}
	}
	return no
}

// sortInt sort the players' troop card.
// They are sorted according to the troop value or in case of a tactic they
// card index.
// #a sorted
func sortInt(a []int) {
	for i := 0; i < len(a); i++ {
		j := i
		for j > 0 && sortInt_Value(a[j-1]) < sortInt_Value(a[j]) {
			a[j], a[j-1] = a[j-1], a[j]
			j--
		}
	}
}

// sortInt_Value calculate the sort value of a card in the players troop.
func sortInt_Value(ix int) int {
	troop, _ := cards.DrTroop(ix)
	if troop != nil {
		return troop.Value()
	} else {
		return ix
	}
}

// eval evaluate a formation.
// #troops is sorted.
func eval(troops []int, env1s []int, env2s []int) (formation *cards.Formation, strenght int) {
	mud, fog := evalEnv(env1s, env2s)
	playedCards := played(troops)
	formation, strenght = evalSim(mud, fog, troops, playedCards)
	return formation, strenght
}

// evalEnv evaluate the environment tactic cards.
func evalEnv(env1s []int, env2s []int) (mud bool, fog bool) {
	envs := make([]int, 4)
	envs = append(envs, env1s...)
	envs = append(envs, env2s...)
	for _, card := range envs {
		if card == cards.TC_Mud {
			mud = true
		} else if card == cards.TC_Fog {
			fog = true
		}
	}
	return mud, fog
}

// evalSim evaluate a formation.
// This function is the part of evaluation that are used for simulation.
// So all calculation, that is possible to do, before a simulation must be done
// before this function is called.
// #troops is sorted.
func evalSim(mud bool, fog bool, troops []int, played int) (formation *cards.Formation, strenght int) {
	sortInt(troops)
	if mud && played == 4 {
		if fog {
			formation = &cards.F_Host
			strenght = evalStrenght(troops, 3, 10)
		} else {
			vformation, v123, vLeader := evalFormation(troops)
			formation = vformation
			strenght = evalStrenght(troops, v123, vLeader)
		}
	} else if played == 3 {
		if fog {
			formation = &cards.F_Host
			strenght = evalStrenght(troops[:3], 3, 10)
		} else {
			vformation, v123, vLeader := evalFormation(troops[:3])
			formation = vformation
			strenght = evalStrenght(troops[:3], v123, vLeader)
		}
	} else {
		formation = nil
		strenght = 0
	}
	return formation, strenght
}

// evalStrenght evalue the strenght of a formation.
func evalStrenght(form []int, v123 int, vLeader int) (st int) {
	for _, cardix := range form {
		troop, _ := cards.DrTroop(cardix)
		if troop != nil {
			st = st + troop.Value()
		} else if cardix == cards.TC_8 {
			st += 8
		} else if cardix == cards.TC_123 {
			st += v123
		} else if cardix == cards.TC_Alexander || cardix == cards.TC_Darius {
			st += vLeader
		} else {
			panic(fmt.Sprintf("Card %v do not exist or is not legal", cardix))
		}
	}
	return st
}

// evalFormation evaluate the formation for a full formation 3 or 4 and no fog.
// Plan split the problem in case of tactic cards.
// Case 1: No tactic card.
// Case 2: One tactic card.
// Case 3: Two tactic cards.
// Case 4: Three tactic cards.
// 3 or 4 cards should not make a difference.
func evalFormation(troopixs []int) (formation *cards.Formation, v123 int, vLeader int) {
	var tac1, tac2, tac3 int
	troops := make([]*cards.Troop, 0, 4)
	troop1, _ := cards.DrTroop(troopixs[0])
	if troop1 == nil {
		tac1 = troopixs[0]
	} else {
		troops = append(troops, troop1)
	}
	troop2, _ := cards.DrTroop(troopixs[1])
	if troop2 == nil {
		tac2 = troopixs[1]
	} else {
		troops = append(troops, troop2)
	}
	troop3, _ := cards.DrTroop(troopixs[2])
	if troop3 == nil {
		tac3 = troopixs[2]
	} else {
		troops = append(troops, troop3)
	}
	if len(troopixs) == 4 {
		troop4, _ := cards.DrTroop(troopixs[3])
		troops = append(troops, troop4)
	}

	if tac3 != 0 {
		formation = &cards.F_BattalionOrder
		v123 = 3
		vLeader = 10
	} else {

		if tac2 != 0 {
			formation, v123, vLeader = evalFormation_T2(troops, tac1, tac2)
		} else if tac1 != 0 {
			formation, v123, vLeader = evalFormation_T1(troops, tac1)
		} else {
			formation, v123, vLeader = evalFormation_T0(troops)
		}
	}

	return formation, v123, vLeader
}

// evalFormation_T1 evalue a formation with one joker.
// troops must be sorted biggest first.
// v123 is the value that the 123 joker takes in the formation.
// vLeader is the value that leader joker takes."
func evalFormation_T1(troops []*cards.Troop, tac int) (formation *cards.Formation, v123 int, vLeader int) {
	value := evalFormation_Value(troops)
	color := evalFormation_Color(troops)
	line, skipValue := evalFormation_T1_Line(troops)
	switch {
	case tac == cards.TC_Alexander || tac == cards.TC_Darius:
		if color {
			if line {
				formation = &cards.F_Wedge
				vLeader = evalFormation_T1_LineLeader(troops, skipValue)
			} else if value {
				formation = &cards.F_Phalanx
				vLeader = troops[0].Value()
			} else {
				formation = &cards.F_BattalionOrder
				vLeader = 10
			}

		} else { // no color
			if value {
				formation = &cards.F_Phalanx
				vLeader = troops[0].Value()
			} else if line {
				formation = &cards.F_SkirmishLine
				vLeader = evalFormation_T1_LineLeader(troops, skipValue)
			} else {
				formation = &cards.F_Host
				vLeader = 10
			}
		}
	case tac == cards.TC_8:
		line8 := false
		if line {
			line8 = evalFormation_T1_8Line(troops, skipValue)
		}
		if color {
			if line8 {
				formation = &cards.F_Wedge
			} else if value && troops[0].Value() == 8 {
				formation = &cards.F_Phalanx
			} else {
				formation = &cards.F_BattalionOrder
			}
		} else { // no color
			if value && troops[0].Value() == 8 {
				formation = &cards.F_Phalanx
			} else if line8 {
				formation = &cards.F_SkirmishLine
			} else {
				formation = &cards.F_Host
			}
		}

	case tac == cards.TC_123:
		line123 := false
		if line {
			line123 = evalFormation_T1_123Line(troops, skipValue)
		}
		if color {
			if line123 {
				formation = &cards.F_Wedge
				v123 = evalFormation_T1_Line123(troops, skipValue)
			} else if value && troops[0].Value() < 4 {
				formation = &cards.F_Phalanx
				v123 = troops[0].Value()
			} else {
				formation = &cards.F_BattalionOrder
				v123 = 3
			}
		} else { // no color
			if value && troops[0].Value() < 4 {
				formation = &cards.F_Phalanx
				v123 = troops[0].Value()
			} else if line123 {
				formation = &cards.F_SkirmishLine
				v123 = evalFormation_T1_Line123(troops, skipValue)
			} else {
				formation = &cards.F_Host
				v123 = 3
			}
		}
	default:
		panic("Unexpected combination of tactic cards")
	}
	return formation, v123, vLeader

}

// evalFormation_T1_8Line check if there is a line.
// troops are trimed.
func evalFormation_T1_8Line(troops []*cards.Troop, skipValue int) (line bool) {
	if skipValue == 0 {
		if troops[0].Value() == 7 || (troops[0].Value() == 10 && len(troops) == 2) {
			line = true
		}
	} else if skipValue == 8 {
		line = true
	}

	return line
}

//evalFormation_T1_123Line check for a line with a 123 joker.
//troops are trimed.
func evalFormation_T1_123Line(troops []*cards.Troop, skipValue int) (line bool) {
	if skipValue == 0 {
		if (troops[len(troops)-1].Value() < 5 && troops[len(troops)-1].Value() != 1) || troops[0].Value() == 2 {
			line = true
		}
	} else if skipValue < 4 { //3 or 2
		line = true
	}

	return line
}

// evalFormation_T1_Line123 evaluate the 123 joker strenght value of a line formation.
// troops must be sorted bigest first.
func evalFormation_T1_Line123(troops []*cards.Troop, skipValue int) (v123 int) {
	if skipValue == 0 {
		switch troops[0].Value() {
		case 5:
			v123 = 3
		case 4:
			v123 = 2
		case 3:
			v123 = 1
		case 2:
			v123 = 3
		default:
			panic("This is no expected 123 line")
		}
	} else {
		v123 = skipValue
	}
	return v123
}

// evalFormation_T1_LineLeader evaluate the leader joker strenght value in a line formation.
// troops must be sorted bigest first.
func evalFormation_T1_LineLeader(troops []*cards.Troop, skipValue int) (leader int) {
	if skipValue == 0 {
		if troops[0].Value() != 10 {
			leader = troops[0].Value() + 1
		} else {
			leader = troops[len(troops)-1].Value() - 1
		}
	} else {
		leader = skipValue
	}
	return leader
}

// evalFormation_T2 evaluate a formation with two jokers.
// troops must be sorted biggest first.
// v123 is the value that the 123 joker takes in the formation.
// vLeader is the value that leader joker takes."
func evalFormation_T2(troops []*cards.Troop, tac1 int, tac2 int) (formation *cards.Formation, v123 int, vLeader int) {
	switch {
	case (tac1 == cards.TC_Alexander || tac1 == cards.TC_Darius) && tac2 == cards.TC_8:
		color := evalFormation_Color(troops)

		if len(troops) == 1 {
			if troops[0].Value() == 10 {
				formation = &cards.F_Wedge
				vLeader = 9
			} else if troops[0].Value() == 9 {
				formation = &cards.F_Wedge
				vLeader = 10
			} else if troops[0].Value() == 8 {
				formation = &cards.F_Phalanx
				vLeader = 8
			} else if troops[0].Value() == 7 {
				formation = &cards.F_Wedge
				vLeader = 9
			} else if troops[0].Value() == 6 {
				formation = &cards.F_Wedge
				vLeader = 7
			} else {
				formation = &cards.F_BattalionOrder
				vLeader = 10
			}

		} else { // Two troops
			if troops[0].Value() == 10 && troops[1].Value() == 9 {
				formation = &cards.F_Wedge
				vLeader = 7
			} else if troops[0].Value() == 10 && troops[1].Value() == 7 {
				formation = &cards.F_Wedge
				vLeader = 9
			} else if troops[0].Value() == 9 && troops[1].Value() == 7 {
				formation = &cards.F_Wedge
				vLeader = 10
			} else if troops[0].Value() == 9 && troops[1].Value() == 6 {
				formation = &cards.F_Wedge
				vLeader = 7
			} else if troops[0].Value() == 7 && troops[1].Value() == 6 {
				formation = &cards.F_Wedge
				vLeader = 9
			} else if troops[0].Value() == 7 && troops[1].Value() == 5 {
				formation = &cards.F_Wedge
				vLeader = 6
			} else if troops[0].Value() == 6 && troops[1].Value() == 5 {
				formation = &cards.F_Wedge
				vLeader = 7
			} else {
				formation = &cards.F_BattalionOrder
				vLeader = 10
			}
			if troops[0].Value() == 8 && troops[1].Value() == 8 {
				formation = &cards.F_Phalanx
				vLeader = 8
			} else {
				if !color {
					if formation == &cards.F_Wedge {
						formation = &cards.F_SkirmishLine
					} else {
						formation = &cards.F_Host
					}
				}
			}
		}

	case (tac1 == cards.TC_Alexander || tac1 == cards.TC_Darius) && tac2 == cards.TC_123:
		if len(troops) == 1 {
			if troops[0].Value() < 6 {
				formation = &cards.F_Wedge
				if troops[0].Value() == 1 {
					v123 = 2
					vLeader = 3
				} else if troops[0].Value() == 2 {
					v123 = 3
					vLeader = 4
				} else if troops[0].Value() == 3 {
					v123 = 2
					vLeader = 4
				} else if troops[0].Value() == 4 {
					v123 = 3
					vLeader = 5
				} else if troops[0].Value() == 5 {
					v123 = 3
					vLeader = 4
				}
			} else {
				formation = &cards.F_BattalionOrder
				vLeader = 10
				v123 = 3
			}
		} else { // two troops
			value := evalFormation_Value(troops)
			color := evalFormation_Color(troops)
			if value && troops[0].Value() < 4 {
				vLeader = troops[0].Value()
				v123 = troops[0].Value()
				formation = &cards.F_Phalanx
			} else if troops[0].Value() == 3 && troops[1].Value() == 2 {
				v123 = 1
				vLeader = 4
			} else if troops[0].Value() == 4 && troops[1].Value() == 1 {
				v123 = 2
				vLeader = 3
			} else if troops[0].Value() == 3 && troops[1].Value() == 1 {
				v123 = 2
				vLeader = 4
			} else if troops[0].Value() == 5 && troops[1].Value() == 3 {
				v123 = 2
				vLeader = 4
			} else if troops[0].Value() == 4 && troops[1].Value() == 3 {
				v123 = 2
				vLeader = 5
			} else if troops[0].Value() == 2 && troops[1].Value() == 1 {
				v123 = 3
				vLeader = 4
			} else if troops[0].Value() == 6 && troops[1].Value() == 5 {
				v123 = 3
				vLeader = 4
			} else if troops[0].Value() == 4 && troops[1].Value() == 2 {
				v123 = 3
				vLeader = 5
			} else if troops[0].Value() == 6 && troops[1].Value() == 4 {
				v123 = 3
				vLeader = 5
			} else if troops[0].Value() == 5 && troops[1].Value() == 4 {
				v123 = 3
				vLeader = 6
			}
			if formation == nil { //no phalanx
				if color {
					if v123 != 0 { //line
						formation = &cards.F_Wedge
					} else {
						formation = &cards.F_BattalionOrder
						v123 = 3
						vLeader = 10
					}
				} else {
					if v123 != 0 { //line
						formation = &cards.F_SkirmishLine
					} else {
						formation = &cards.F_Host
						v123 = 3
						vLeader = 10
					}
				}
			}
		}
	case tac1 == cards.TC_8 && tac2 == cards.TC_123:
		color := evalFormation_Color(troops)
		if color {
			formation = &cards.F_BattalionOrder
			v123 = 3
		} else { // no color
			formation = &cards.F_Host
			v123 = 3
		}
	default:
		panic("Unexpected combination of tactic cards")
	}
	return formation, v123, vLeader
}

// evalFormation_T0 evaluate a formation with zero jokers.
// troops must be sorted.
// v123 is the value that the 123 joker takes in the formation.
// vLeader is the value that leader joker takes."
func evalFormation_T0(troops []*cards.Troop) (formation *cards.Formation, v123 int, vLeader int) {
	v123 = 0    //always zero
	vLeader = 0 //always zero
	color := evalFormation_Color(troops)
	if color {
		formation = &cards.F_BattalionOrder
	}
	line := evalFormation_Line(troops)
	if formation != nil { // battalion or wedge
		if line {
			formation = &cards.F_Wedge
		}
	} else { //no battalion or wedge
		value := evalFormation_Value(troops)
		if value {
			formation = &cards.F_Phalanx
		}
		if formation == nil { // no battalion, wedge or phalanx
			if line {
				formation = &cards.F_SkirmishLine
			} else { // no battalion wedge  phalanx Line
				formation = &cards.F_Host
			}
		}
	}
	return formation, v123, vLeader
}
func evalFormation_Color(troops []*cards.Troop) (color bool) {
	color = true
	for i, v := range troops {
		if i != len(troops)-1 {
			if v.Color() != troops[i+1].Color() {
				color = false
				break
			}
		}
	}
	return color
}
func evalFormation_Value(troops []*cards.Troop) (value bool) {
	value = true
	for i, v := range troops {
		if i != len(troops)-1 {
			if v.Value() != troops[i+1].Value() {
				value = false
				break
			}
		}
	}
	return value
}

// evalFormation_Line evalute a line assums the cards are sorted bigest first.
func evalFormation_Line(troops []*cards.Troop) (line bool) {
	line = true
	for i, v := range troops {
		if i != len(troops)-1 {
			if v.Value() != troops[i+1].Value()+1 {
				line = false
				break
			}
		}
	}
	return line
}

// evalFormation_T1_Line evaluate a line with one joker.
// Expect troops to sorted biggest first and trimmed no nil values
// The skipValue is the jokers strenght if zero it is not determent,
func evalFormation_T1_Line(troops []*cards.Troop) (line bool, skipValue int) {
	line = true
	for i, v := range troops {
		if i != len(troops)-1 {
			if v.Value() != troops[i+1].Value()+1 {
				if v.Value() == troops[i+1].Value()+2 && skipValue == 0 {
					skipValue = v.Value() - 1
				} else {
					line = false
					break
				}
			}
		}
	}
	return line, skipValue
}

// ClaimFlag claims a flag if possible.
// Its only possible to claim a flag if player have a formation, if
// both players have a formation it is easy, or if player have the highest formation.
// The main task is calculate the highest possible formation.
// I do not have a plan yet but I think we need brute force, but
// If we are missing four cards and only one combi exist, it's 53*52*51*54=0/24 permutaitions
// ca 300.000 and how do generate them.
// The player's formation do limit the possible formation to the ones that are better but again
// that may be many all to be exact, but if we find one that is better we can stop.
// I see 3 option:
//      1) Just try all permution and if one is higher stop.
//      2) For every combination try to make one with the remaining cards
//      3) A combination a quick check if player have two troops many combination
//         can be ruled out. color, value and bigger bridge than 2 no line.
//         These three check give a max combination.
//         ranging from top to button
//
// Made permutaitor and will try that first it may be quick ennough.
// Have the plan for the second solution. Sort the remaining cards a value/card map per color and
// a value slice of cards map. Then a switch with drop through on the combination and each combination
// function use the sorted cards
func (flag *Flag) ClaimFlag(playerix int, unPlayCards []int) (ok bool, eks []int) {
	opPlayer := flag.Players[opponent(playerix)]
	player := flag.Players[playerix] // updated
	mud, fog := evalEnv(player.Env[:], opPlayer.Env[:])
	if !opPlayer.Won || !player.Won {
		if player.Strenght != 0 { // Formation
			if opPlayer.Strenght != 0 {
				if player.Formation.Value > opPlayer.Formation.Value {
					ok = true
				} else if player.Formation.Value == opPlayer.Formation.Value &&
					player.Strenght > opPlayer.Strenght {
					ok = true
				} else {
					ok = false
					copy(eks, opPlayer.Troops[:])
				}
			} else { // opponent no formation
				sortInt(opPlayer.Troops[:]) //============Sort Oponnent Troops=====================
				opTroops := make([]*cards.Troop, 0, 3)
				for _, v := range opPlayer.Troops {
					opTroop, _ := cards.DrTroop(v)
					if opTroop != nil {
						opTroops = append(opTroops, opTroop)
					}
				}
				if len(opTroops) > 1 {
					color := evalFormation_Color(opTroops)
					value := evalFormation_Value(opTroops)
					line, _ := evalFormation_T1_Line(opTroops) // could be better but for now ok. must be sorted
					wedge := line && color
					switch player.Formation {
					case &cards.F_Wedge:
						if !wedge {
							ok = true
						}
					case &cards.F_Phalanx:
						if !wedge && !value {
							ok = true
						} else {
							m := 3
							if mud {
								m = 4
							}
							if value && opTroops[0].Value()*m < player.Strenght {
								ok = true
							}
						}
					case &cards.F_BattalionOrder:
						if !wedge && !value && !color {
							ok = true
						}
					case &cards.F_SkirmishLine:
						if !color && !value && !line {
							ok = true
						}
					}
				}

				if !ok {
					playedCards := played(opPlayer.Troops[:])
					opCards := make([]int, 4)
					copy(opCards, opPlayer.Troops[:])
					sortInt(opCards)
					simNoCards := 3 - playedCards
					simPlayedCards := 3
					if mud {
						simNoCards = simNoCards + 1
						simPlayedCards = 4
					}
					simCards := make([]int, 4)
					switch simNoCards {
					case 1:
						for _, card := range unPlayCards {
							copy(simCards, opCards)
							simCards[len(opCards)-1] = card
							simFormation, simStrenght := evalSim(mud, fog, simCards, simPlayedCards)
							if simFormation.Value > player.Formation.Value ||
								(simFormation.Value == player.Formation.Value && simStrenght > player.Strenght) {
								eks = simCards
								ok = false
								break
							}
						}
					case 2:
						match := Math.Perm2(len(unPlayCards), func(v [2]int) bool {
							copy(simCards, opCards)
							simCards[len(opCards)-1] = unPlayCards[v[1]]
							simCards[len(opCards)-2] = unPlayCards[v[0]]
							simFormation, simStrenght := evalSim(mud, fog, simCards, simPlayedCards)
							if simFormation.Value > player.Formation.Value {
								eks = simCards
								return true
							} else if simFormation.Value == player.Formation.Value && simStrenght > player.Strenght {
								return true
							} else {
								return false
							}
						})
						if match[0] == -1 {
							ok = true
						}
					case 3:
						match := Math.Perm3(len(unPlayCards), func(v [3]int) bool {
							copy(simCards, opCards)
							simCards[len(opCards)-1] = unPlayCards[v[2]]
							simCards[len(opCards)-2] = unPlayCards[v[1]]
							simCards[len(opCards)-3] = unPlayCards[v[0]]
							simFormation, simStrenght := evalSim(mud, fog, simCards, simPlayedCards)
							if simFormation.Value > player.Formation.Value {
								eks = simCards
								return true
							} else if simFormation.Value == player.Formation.Value && simStrenght > player.Strenght {
								return true
							} else {
								return false
							}
						})
						if match[0] == -1 {
							ok = true
						}
					case 4:
						match := Math.Perm4(len(unPlayCards), func(v [4]int) bool {
							copy(simCards, opCards)
							simCards[len(opCards)-1] = unPlayCards[v[3]]
							simCards[len(opCards)-2] = unPlayCards[v[2]]
							simCards[len(opCards)-3] = unPlayCards[v[1]]
							simCards[len(opCards)-4] = unPlayCards[v[0]]
							simFormation, simStrenght := evalSim(mud, fog, simCards, simPlayedCards)
							if simFormation.Value > player.Formation.Value {
								eks = simCards
								return true
							} else if simFormation.Value == player.Formation.Value && simStrenght > player.Strenght {
								eks = simCards
								return true
							} else {
								return false
							}
						})
						if match[0] == -1 {
							ok = true
						}
					}

				}
			}
		}
	}
	if ok {
		player.Won = true
	}
	return ok, eks
}
func (flag *Flag) UsedTac() (v [2][]int) {
	v[0] = usedTac_Tac(flag.Players[0].Env[:], flag.Players[0].Troops[:])
	v[1] = usedTac_Tac(flag.Players[1].Env[:], flag.Players[1].Troops[:])
	return v
}
func usedTac_Tac(Env []int, troops []int) (tacs []int) {
	tacs = make([]int, 0, 5)
	for _, e := range Env {
		if e != 0 {
			tacs = append(tacs, e)
		}
	}
	for _, cardix := range troops {
		if cardix != 0 {
			_, err := cards.DrTactic(cardix)
			if err != nil {
				tacs = append(tacs, cardix)
			}
		}
	}
	return tacs
}
func (flag *Flag) Free() (v [2]bool) {
	mud, _ := evalEnv(flag.Players[0].Env[:], flag.Players[1].Env[:])
	p1 := played(flag.Players[0].Troops[:])
	p2 := played(flag.Players[1].Troops[:])
	space := 3
	if mud {
		space = 4
	}
	v[0] = space-p1 > 0
	v[1] = space-p2 > 0
	return v
}
func (flag *Flag) Env(pix int) (copyenv []int) {
	env := flag.Players[pix].Env[:]
	copyenv = make([]int, 0, len(env))
	for _, v := range env {
		if v != 0 {
			copyenv = append(copyenv, v)
		}
	}
	return copyenv
}
func (flag *Flag) Troops(pix int) (copyTroops []int) {
	troops := flag.Players[pix].Troops[:]
	copyTroops = make([]int, 0, len(troops))
	for _, v := range troops {
		if v != 0 {
			copyTroops = append(copyTroops, v)
		}
	}
	return copyTroops
}
func (flag *Flag) Claimed() bool {
	return flag.Players[0].Won || flag.Players[1].Won
}
func (flag *Flag) Won() (res [2]bool) {
	res[0] = flag.Players[0].Won
	res[1] = flag.Players[1].Won
	return res
}
func (flag *Flag) Formations() (form [2]*cards.Formation) {
	form[0] = flag.Players[0].Formation
	form[1] = flag.Players[1].Formation
	return form
}
