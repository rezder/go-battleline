package flag

import (
	"errors"
	"fmt"
	"rezder.com/game/card/battleline/cards"
	Math "rezder.com/math/int"
	"sort"
)

const (
	vm123    = 3
	vmLeader = 10
)

type Player struct {
	won       bool
	env       [2]int
	troops    [4]int
	formation *cards.Formation
	strenght  int
}

type Flag struct {
	players [2]Player
}

func (flag *Flag) Remove(cardix int, playerix int) (mudix0 int, mudix1 int, err error) {
	mudix1 = -1
	mudix0 = -1
	card, err := cards.DrCard(cardix)
	if err == nil {
		errMessage := fmt.Sprintf("Player %v do not have card %v", playerix, card.Name())
		switch tcard := card.(type) {
		case cards.Tactic:
			if tcard.Type() == cards.T_Env {
				delix := remove_clear(flag.players[playerix].env[:], cardix)
				if delix != -1 {
					if cardix == cards.TC_Mud && played(flag.players[1].troops[:]) == 4 || played(flag.players[0].troops[:]) == 4 {
						if played(flag.players[0].troops[:]) == 4 {
							mudix0 = flag.remove_Mud(0)
							remove_clear(flag.players[0].troops[:], mudix0)
						}
						if played(flag.players[1].troops[:]) == 4 {
							mudix1 = flag.remove_Mud(1)
							remove_clear(flag.players[1].troops[:], mudix1)
						}
					}
					flag.updateFormations()
				} else {
					err = errors.New(errMessage)
				}
			} else if tcard.Type() == cards.T_Morale {
				flag.remove_Troop(cardix, playerix, errMessage)
			} else {
				err = errors.New(fmt.Sprintf("Illegal tactic card: %v", card.Name()))
			}
		case cards.Troop:
			flag.remove_Troop(cardix, playerix, errMessage)
		default:
			panic("No supprted type")
		}
	} else {
		panic("Card do not exist")
	}
	return mudix0, mudix1, err
}

// remove_Mud find the excess card that gives the highest formation
// and strenght when it is removed.
func (flag *Flag) remove_Mud(playerix int) (mudix int) {
	var maxf *cards.Formation = &cards.F_Host
	var maxs int = 1
	var troops []int
	for i, v := range flag.players[playerix].troops {
		troops = make([]int, 4)
		copy(troops, flag.players[playerix].troops[:])
		troops[i] = 0
		formation, strength := eval(troops, flag.players[playerix].env[:], flag.players[opponent(playerix)].env[:])
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
func remove_clear(cards []int, cardix int) (ix int) {
	ix = -1
	for i, v := range cards {
		if v == cardix {
			cards[i] = 0
			ix = i
			break
		}
	}
	return ix
}
func (flag *Flag) remove_Troop(cardix int, playerix int, errM string) (err error) {
	place := remove_clear(flag.players[playerix].troops[:], cardix)
	if place != -1 {
		flag.updateFormation(playerix)
	} else {
		err = errors.New("errM")
	}
	return err
}
func opponent(playerix int) int {
	if playerix == 0 {
		return 1
	} else {
		return 0
	}
}
func (flag *Flag) Set(cardix int, playerix int) (err error) {
	card, err := cards.DrCard(cardix)
	if err == nil {
		switch tcard := card.(type) {
		case cards.Tactic:
			if tcard.Type() == cards.T_Env {
				place := set_Card(cardix, flag.players[playerix].env[:])
				if place != -1 {
					flag.updateFormations()
				} else {
					err = errors.New("No available slots")
				}
			} else if tcard.Type() == cards.T_Morale {
				flag.set_Troop(cardix, playerix)
			} else {
				err = errors.New("Illegal tactic card: " + card.Name())
			}
		case cards.Troop:
			flag.set_Troop(cardix, playerix)
		default:
			fmt.Printf("card type %v", tcard)
			panic("No supprted type")
		}
	} else {
		panic("Card do not exist")
	}
	return err
}
func set_Card(cardix int, v []int) (place int) {
	place = -1
	for i, c := range v {
		if c == 0 {
			v[i] = cardix
			place = i
			break
		}
	}
	return place
}
func (flag *Flag) set_Troop(cardix int, playerix int) (err error) {
	place := set_Card(cardix, flag.players[playerix].troops[:])
	if place != -1 {
		flag.updateFormation(playerix)
	} else {
		err = errors.New("No available slots")
	}
	return err
}
func (flag *Flag) updateFormations() {
	flag.updateFormation(0)
	flag.updateFormation(1)
}
func played(troops []int) (no int) {
	no = 0
	for _, troop := range troops {
		if troop != 0 {
			no++
		}
	}
	return no
}
func (flag *Flag) updateFormation(playerix int) {
	player := &flag.players[playerix]
	troops := player.troops[:]
	envs := player.env[:]
	formation, strenght := eval(troops, envs, flag.players[opponent(playerix)].env[:])
	player.formation = formation
	player.strenght = strenght

}
func sortInt(a []int) {
	for i := 0; i < len(a); i++ {
		j := i
		for j > 0 && sortInt_Value(a[j-1]) < sortInt_Value(a[j]) {
			a[j], a[j-1] = a[j-1], a[j]
			j--
		}
	}
}

func sortInt_Value(ix int) int {
	troop, _ := cards.DrTroop(ix)
	if troop != nil {
		return troop.Value()
	} else {
		return ix
	}
}

//eval evaluate a formation.
//Warning only the first two slices is sorted, so they should be the one from the changed formation.
func eval(troops []int, env1s []int, env2s []int) (formation *cards.Formation, strenght int) {
	mud, fog := evalEnv(env1s, env2s)
	playedCards := played(troops)
	formation, strenght = evalSim(mud, fog, troops, playedCards)
	return formation, strenght
}
func evalEnv(env1s []int, env2s []int) (mud bool, fog bool) {
	sort.Sort(sort.Reverse(sort.IntSlice(env1s)))
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

//evalFormation evaluate the formation for a full formation 3 or 4 and no fog.
//Plan split the problem in case of tactic cards.
//Case 1: No tactic card
//Case 2: One tactic card
//Case 3: Two tactic cards
//Case 4: Three tactic cards
//3 or 4 cards should not make a difference
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
func evalFormation_T1(troops []*cards.Troop, tac int) (formation *cards.Formation, v123 int, vLeader int) {
	value := evalFormation_Value(troops)
	color := evalFormation_Color(troops)
	line, skipValue := evalFormation_T1_Line(troops)
	switch {
	case tac == cards.TC_Alexander || tac == cards.TC_Darius:
		if color {
			if line {
				formation = &cards.F_Wedge
				vLeader = evalFormation_T1_Leader(troops, skipValue)
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
				vLeader = evalFormation_T1_Leader(troops, skipValue)
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
				v123 = evalFormation_T1_V123(troops, skipValue)
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
				v123 = evalFormation_T1_V123(troops, skipValue)
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
func evalFormation_T1_V123(troops []*cards.Troop, skipValue int) (v123 int) {
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
func evalFormation_T1_Leader(troops []*cards.Troop, skipValue int) (leader int) {
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

// ClaimFlag Claims a flag if possible.
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
	opPlayer := flag.players[opponent(playerix)]
	mud, fog := evalEnv(flag.players[playerix].env[:], opPlayer.env[:])
	if !opPlayer.won || !flag.players[playerix].won {
		if flag.players[playerix].strenght != 0 { // Formation
			if opPlayer.strenght != 0 {
				if flag.players[playerix].formation.Value > opPlayer.formation.Value {
					ok = true
				} else if flag.players[playerix].formation.Value == opPlayer.formation.Value &&
					flag.players[playerix].strenght > opPlayer.strenght {
					ok = true
				} else {
					ok = false
					copy(eks, opPlayer.troops[:])
				}
			} else { // opponent no formation
				sortInt(opPlayer.troops[:]) //============Sort Oponnent Troops=====================
				opTroops := make([]*cards.Troop, 0, 3)
				for _, v := range opPlayer.troops {
					opTroop, _ := cards.DrTroop(v)
					if opTroop != nil {
						opTroops = append(opTroops, opTroop)
					}
				}
				formation := flag.players[playerix].formation
				strenght := flag.players[playerix].strenght
				if len(opTroops) > 1 {
					color := evalFormation_Color(opTroops)
					value := evalFormation_Value(opTroops)
					line, _ := evalFormation_T1_Line(opTroops) // could be better but for now ok. must be sorted
					wedge := line && color
					switch formation {
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
							if value && opTroops[0].Value()*m < flag.players[playerix].strenght {
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
					playedCards := played(opPlayer.troops[:])
					opCards := make([]int, 4)
					copy(opCards, opPlayer.troops[:])
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
							if simFormation.Value > formation.Value ||
								(simFormation.Value == formation.Value && simStrenght > strenght) {
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
							if simFormation.Value > formation.Value {
								eks = simCards
								return true
							} else if simFormation.Value == formation.Value && simStrenght > strenght {
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
							if simFormation.Value > formation.Value {
								eks = simCards
								return true
							} else if simFormation.Value == formation.Value && simStrenght > strenght {
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
							if simFormation.Value > formation.Value {
								eks = simCards
								return true
							} else if simFormation.Value == formation.Value && simStrenght > strenght {
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
		flag.players[playerix].won = true
	}
	return ok, eks
}
