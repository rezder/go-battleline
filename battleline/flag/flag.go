//flag contains a battleline flag.
package flag

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/battleline/cards"
	math "github.com/rezder/go-math/int"
	slice "github.com/rezder/go-slice/int"
	"sort"
)

//Player the structer for player details of the flag.
type Player struct {
	Won       bool
	Env       [2]int
	Troops    [4]int
	Formation *cards.Formation
	Strenght  int
}

//Equal test for equal.
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
func (player *Player) String() (txt string) {
	flag := ""
	if player.Won {
		flag = "#"
	}
	formation := "nil"
	if player.Formation != nil {
		formation = player.Formation.Name
	}
	txt = fmt.Sprintf("Player{%v,%v,%v,%v,%v}", flag, player.Env, player.Troops, formation, player.Strenght)
	return txt
}

//Flag the structer of the flag.
type Flag struct {
	Players [2]*Player
}

//New creates a flag.
func New() (flag *Flag) {
	flag = new(Flag)
	flag.Players[0] = new(Player)
	flag.Players[1] = new(Player)
	return flag
}

//Copy copy a flag.
func (flag *Flag) Copy() (c *Flag) {
	c = new(Flag)
	cp := *flag.Players[0]
	c.Players[0] = &cp
	cp1 := *flag.Players[1]
	c.Players[1] = &cp1

	return c
}

//Equal tests for equal.
func (flag *Flag) Equal(o *Flag) (equal bool) {
	if o == nil && flag == nil {
		equal = true
	} else if o != nil && flag != nil {
		if o == flag {
			equal = true
		} else {
			equal = true
			for i, v := range o.Players {
				if !v.Equal(flag.Players[i]) {
					equal = false
					break
				}
			}
		}
	}
	return equal
}
func (flag *Flag) String() (txt string) {
	if flag != nil {
		flagSign := "#"
		var flagx, flag1, flag2 string
		if flag.Players[0].Won {
			flag1 = flagSign
			flagx = " "
			flag2 = flagx
		} else if flag.Players[1].Won {
			flag2 = flagSign
			flagx = " "
			flag1 = flagx
		} else {
			flag1 = " "
			flagx = flagSign
			flag2 = flag1
		}

		var troops [2][]int
		var envs [2][]string
		for i, p := range flag.Players {
			troop := make([]int, 0, 4)
			for _, t := range p.Troops {
				if t != 0 {
					troop = append(troop, t)
				}
			}
			troops[i] = troop
			env := make([]string, 0, 2)
			var tac *cards.Tactic
			for _, e := range p.Env {
				if e != 0 {
					tac, _ = cards.DrTactic(e)
					env = append(env, tac.Name())
				}
			}
			envs[i] = env
		}
		txt = fmt.Sprintf("Flag{%v %v %v %v %v %v %v }", flag1, troops[0], envs[0], flagx, envs[1], troops[1], flag2)
	} else {
		txt = "Flag{nil}"
	}
	return txt
}

// Remove removes a card from the flag.
// mudix0 contains a card if removal of the mud card result in an excess card for player 0/1
func (flag *Flag) Remove(cardix int, playerix int) (mudix0 int, mudix1 int, err error) {
	player := flag.Players[playerix]        // updated
	opp := flag.Players[opponent(playerix)] // updated
	mudix1 = -1
	mudix0 = -1
	card, ok := cards.DrCard(cardix)
	if !ok {
		panic("Card do not exist")
	}
	errMessage := fmt.Sprintf("Player %v do not have card %v", playerix, card.Name())
	var delix int
	switch tcard := card.(type) {
	case cards.Tactic:
		if tcard.Type() == cards.CTEnv {
			delix, err = removeClear(player.Env[:], cardix, errMessage)
			if delix != -1 {
				if cardix == cards.TCMud {
					if played(player.Troops[:]) == 4 {
						mudix0 = removeMud(player, opp)
						removeClear(player.Troops[:], mudix0, errMessage)
					}
					if played(opp.Troops[:]) == 4 {
						mudix1 = removeMud(opp, player)
						removeClear(opp.Troops[:], mudix1, errMessage)
					}
				}
				sort.Sort(sort.Reverse(sort.IntSlice(player.Env[:])))
				player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
				opp.Formation, opp.Strenght = eval(opp.Troops[:], opp.Env[:], player.Env[:])
			}
		} else if tcard.Type() == cards.CTMorale {
			delix, err = removeClear(player.Troops[:], cardix, errMessage)
			if delix != -1 {
				player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
			}
		} else {
			err = fmt.Errorf("Illegal tactic card: %v", card.Name())
		}
	case cards.Troop:
		delix, err = removeClear(player.Troops[:], cardix, errMessage)
		if delix != -1 {
			player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
		}

	default:
		panic("Not a supported type")
	}
	return mudix0, mudix1, err
}

// removeMud finds the excess card that gives the highest formation
// and strenght when it is removed.
// There are a problem with this, a dished card can not be a traitor, so removing
// the card that gives the best formation may not be the best move. The rules does not
// cover this.
func removeMud(player *Player, opp *Player) (mudix int) {
	maxf := &cards.FHost
	maxs := 1
	var troops []int
	for i, v := range player.Troops {
		troops = make([]int, 4) //copy
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

// removeClear zero a card index.
// ix the index in list if the card was found if not -1.
func removeClear(cards []int, cardix int, errM string) (ix int, err error) {
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
	}
	return 0
}

//Set a card.
func (flag *Flag) Set(cardix int, playerix int) (err error) {
	player := flag.Players[playerix]        // updated
	opp := flag.Players[opponent(playerix)] // updated
	card, ok := cards.DrCard(cardix)
	if !ok {
		panic("Card do not exist")
	}
	errMessage := fmt.Sprintf("Player %v do not have space for card %v", playerix, card.Name())
	var place int
	switch tcard := card.(type) {
	case cards.Tactic:
		if tcard.Type() == cards.CTEnv {
			place, err = setCard(cardix, player.Env[:], errMessage)
			if place != -1 {
				player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
				opp.Formation, opp.Strenght = eval(opp.Troops[:], opp.Env[:], player.Env[:])
			}
		} else if tcard.Type() == cards.CTMorale {
			place, err = setCard(cardix, player.Troops[:], errMessage)
			if place != -1 {
				player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
			}
		} else {
			err = errors.New("Illegal tactic card: " + card.Name())
		}
	case cards.Troop:
		place, err = setCard(cardix, player.Troops[:], errMessage)
		if place != -1 {
			player.Formation, player.Strenght = eval(player.Troops[:], player.Env[:], opp.Env[:])
		}
	default:
		txt := fmt.Sprintf("Card type: %v not supported.", tcard)
		panic(txt)
	}
	return err
}

// set_Card set a card in a list of cards.
// #v set a card at the first available spot.
func setCard(cardix int, v []int, errM string) (place int, err error) {
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
		for j > 0 && sortIntValue(a[j-1]) < sortIntValue(a[j]) {
			a[j], a[j-1] = a[j-1], a[j]
			j--
		}
	}
}

// sortIntValue calculate the sort value of a card in the players troop.
func sortIntValue(ix int) int {
	troop, ok := cards.DrTroop(ix)
	if ok {
		return troop.Value()
	}
	return ix
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
	envs := make([]int, 0, 4)
	envs = append(envs, env1s...)
	envs = append(envs, env2s...)
	for _, card := range envs {
		if card == cards.TCMud {
			mud = true
		} else if card == cards.TCFog {
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
			formation = &cards.FHost
			strenght = evalStrenght(troops, 3, 10)
		} else {
			vformation, v123, vLeader := evalFormation(troops)
			formation = vformation
			strenght = evalStrenght(troops, v123, vLeader)
		}
	} else if played == 3 && !mud {
		if fog {
			formation = &cards.FHost
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
		troop, ok := cards.DrTroop(cardix)
		if ok {
			st = st + troop.Value()
		} else if cardix == cards.TC8 {
			st += 8
		} else if cardix == cards.TC123 {
			st += v123
		} else if cardix == cards.TCAlexander || cardix == cards.TCDarius {
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
	troop, ok := cards.DrTroop(troopixs[0])
	if !ok {
		tac1 = troopixs[0]
	} else {
		troops = append(troops, troop)
	}
	troop, ok = cards.DrTroop(troopixs[1])
	if !ok {
		tac2 = troopixs[1]
	} else {
		troops = append(troops, troop)
	}
	troop, ok = cards.DrTroop(troopixs[2])
	if !ok {
		tac3 = troopixs[2]
	} else {
		troops = append(troops, troop)
	}
	if len(troopixs) == 4 { // last card can only be troop as tactic is sorted first.
		troop, _ = cards.DrTroop(troopixs[3])
		troops = append(troops, troop)
	}

	if tac3 != 0 {
		formation = &cards.FBattalion
		v123 = 3
		vLeader = 10
	} else {

		if tac2 != 0 {
			formation, v123, vLeader = evalFormationT2(troops, tac1, tac2)
		} else if tac1 != 0 {
			formation, v123, vLeader = evalFormationT1(troops, tac1)
		} else {
			formation, v123, vLeader = evalFormationT0(troops)
		}
	}

	return formation, v123, vLeader
}

// evalFormationT1 evalue a formation with one joker.
// troops must be sorted biggest first.
// v123 is the value that the 123 joker takes in the formation.
// vLeader is the value that leader joker takes."
func evalFormationT1(troops []*cards.Troop, tac int) (formation *cards.Formation, v123 int, vLeader int) {
	value, color := evalFormationValueColor(troops)
	line, skipValue := evalFormationT1Line(troops)
	switch {
	case tac == cards.TCAlexander || tac == cards.TCDarius:
		if color {
			if line {
				formation = &cards.FWedge
				vLeader = evalFormationT1LineLeader(troops, skipValue)
			} else if value {
				formation = &cards.FPhalanx
				vLeader = troops[0].Value()
			} else {
				formation = &cards.FBattalion
				vLeader = 10
			}

		} else { // no color
			if value {
				formation = &cards.FPhalanx
				vLeader = troops[0].Value()
			} else if line {
				formation = &cards.FSkirmish
				vLeader = evalFormationT1LineLeader(troops, skipValue)
			} else {
				formation = &cards.FHost
				vLeader = 10
			}
		}
	case tac == cards.TC8:
		line8 := false
		if line {
			line8 = evalFormationT18Line(troops, skipValue)
		}
		if color {
			if line8 {
				formation = &cards.FWedge
			} else if value && troops[0].Value() == 8 {
				formation = &cards.FPhalanx
			} else {
				formation = &cards.FBattalion
			}
		} else { // no color
			if value && troops[0].Value() == 8 {
				formation = &cards.FPhalanx
			} else if line8 {
				formation = &cards.FSkirmish
			} else {
				formation = &cards.FHost
			}
		}

	case tac == cards.TC123:
		line123 := false
		if line {
			line123 = evalFormationT1123Line(troops, skipValue)
		}
		if color {
			if line123 {
				formation = &cards.FWedge
				v123 = evalFormationT1Line123(troops, skipValue)
			} else if value && troops[0].Value() < 4 {
				formation = &cards.FPhalanx
				v123 = troops[0].Value()
			} else {
				formation = &cards.FBattalion
				v123 = 3
			}
		} else { // no color
			if value && troops[0].Value() < 4 {
				formation = &cards.FPhalanx
				v123 = troops[0].Value()
			} else if line123 {
				formation = &cards.FSkirmish
				v123 = evalFormationT1Line123(troops, skipValue)
			} else {
				formation = &cards.FHost
				v123 = 3
			}
		}
	default:
		panic("Unexpected combination of tactic cards")
	}
	return formation, v123, vLeader

}

// evalFormationT18Line check if there is a line.
// troops are trimed.
func evalFormationT18Line(troops []*cards.Troop, skipValue int) (line bool) {
	if skipValue == 0 {
		if troops[0].Value() == 7 || (troops[0].Value() == 10 && len(troops) == 2) {
			line = true
		}
	} else if skipValue == 8 {
		line = true
	}

	return line
}

//evalFormationT1123Line check for a line with a 123 joker.
//troops are trimed.
func evalFormationT1123Line(troops []*cards.Troop, skipValue int) (line bool) {
	if skipValue == 0 {
		if (troops[len(troops)-1].Value() < 5 && troops[len(troops)-1].Value() != 1) || troops[0].Value() == 2 {
			line = true
		}
	} else if skipValue < 4 { //3 or 2
		line = true
	}

	return line
}

// evalFormationT1Line123 evaluate the 123 joker strenght value of a line formation.
// troops must be sorted bigest first.
func evalFormationT1Line123(troops []*cards.Troop, skipValue int) (v123 int) {
	if skipValue == 0 {
		switch troops[0].Value() {
		case 6:
			v123 = 3
		case 5:
			v123 = 5 - len(troops)
		case 4:
			v123 = 4 - len(troops)
		case 3:
			v123 = 1
		case 2:
			v123 = 3
		default:
			panic(fmt.Sprintf("This is no expected 123 line first troop value: %v, troops: ", troops[0].Value(), troops))
		}
	} else {
		v123 = skipValue
	}
	return v123
}

// evalFormationT1LineLeader evaluate the leader joker strenght value in a line formation.
// troops must be sorted bigest first.
func evalFormationT1LineLeader(troops []*cards.Troop, skipValue int) (leader int) {
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

//evalFormationLeader8 find the formation and the leader value for a flag
//with a leader and the 8 morale tactic card.
//troops must be sorted biggest first.
//vLeader is the value that leader takes."
func evalFormationLeader8(troops []*cards.Troop) (formation *cards.Formation, vLeader int) {
	_, color := evalFormationValueColor(troops)

	if len(troops) == 1 {
		if troops[0].Value() == 10 {
			formation = &cards.FWedge
			vLeader = 9
		} else if troops[0].Value() == 9 {
			formation = &cards.FWedge
			vLeader = 10
		} else if troops[0].Value() == 8 {
			formation = &cards.FPhalanx
			vLeader = 8
		} else if troops[0].Value() == 7 {
			formation = &cards.FWedge
			vLeader = 9
		} else if troops[0].Value() == 6 {
			formation = &cards.FWedge
			vLeader = 7
		} else {
			formation = &cards.FBattalion
			vLeader = 10
		}

	} else { // Two troops
		if troops[0].Value() == 10 && troops[1].Value() == 9 {
			formation = &cards.FWedge
			vLeader = 7
		} else if troops[0].Value() == 10 && troops[1].Value() == 7 {
			formation = &cards.FWedge
			vLeader = 9
		} else if troops[0].Value() == 9 && troops[1].Value() == 7 {
			formation = &cards.FWedge
			vLeader = 10
		} else if troops[0].Value() == 9 && troops[1].Value() == 6 {
			formation = &cards.FWedge
			vLeader = 7
		} else if troops[0].Value() == 7 && troops[1].Value() == 6 {
			formation = &cards.FWedge
			vLeader = 9
		} else if troops[0].Value() == 7 && troops[1].Value() == 5 {
			formation = &cards.FWedge
			vLeader = 6
		} else if troops[0].Value() == 6 && troops[1].Value() == 5 {
			formation = &cards.FWedge
			vLeader = 7
		} else {
			formation = &cards.FBattalion
			vLeader = 10
		}
		if troops[0].Value() == 8 && troops[1].Value() == 8 {
			formation = &cards.FPhalanx
			vLeader = 8
		} else {
			if !color {
				if formation == &cards.FWedge {
					formation = &cards.FSkirmish
				} else {
					formation = &cards.FHost
				}
			}
		}
	}
	return formation, vLeader
}

//evalFormationLeader123 finds a formation and the leader value for a flag
//with a leader and the 123 morale tactic card.
//troops must be sorted biggest first.
//vLeader is the value that leader takes."
//v123 is the value 123 takes
func evalFormationLeader123(troops []*cards.Troop) (formation *cards.Formation, v123, vLeader int) {
	if len(troops) == 1 {
		if troops[0].Value() < 6 {
			formation = &cards.FWedge
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
			formation = &cards.FBattalion
			vLeader = 10
			v123 = 3
		}
	} else { // two troops
		value, color := evalFormationValueColor(troops)
		if value && troops[0].Value() < 4 {
			vLeader = troops[0].Value()
			v123 = troops[0].Value()
			formation = &cards.FPhalanx
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
					formation = &cards.FWedge
				} else {
					formation = &cards.FBattalion
					v123 = 3
					vLeader = 10
				}
			} else {
				if v123 != 0 { //line
					formation = &cards.FSkirmish
				} else {
					formation = &cards.FHost
					v123 = 3
					vLeader = 10
				}
			}
		}
	}
	return formation, v123, vLeader
}

// evalFormationT2 evaluate a formation with two jokers.
// troops must be sorted biggest first.
// v123 is the value that the 123 joker takes in the formation.
// vLeader is the value that leader joker takes."
func evalFormationT2(troops []*cards.Troop, tac1 int, tac2 int) (formation *cards.Formation, v123 int, vLeader int) {
	switch {
	case (tac1 == cards.TCAlexander || tac1 == cards.TCDarius) && tac2 == cards.TC8:
		formation, vLeader = evalFormationLeader8(troops)

	case (tac1 == cards.TCAlexander || tac1 == cards.TCDarius) && tac2 == cards.TC123:
		formation, v123, vLeader = evalFormationLeader123(troops)
	case tac1 == cards.TC8 && tac2 == cards.TC123:
		_, color := evalFormationValueColor(troops)
		if color {
			formation = &cards.FBattalion
			v123 = 3
		} else { // no color
			formation = &cards.FHost
			v123 = 3
		}
	default:
		panic("Unexpected combination of tactic cards")
	}
	return formation, v123, vLeader
}

// evalFormationT0 evaluate a formation with zero jokers.
// troops must be sorted.
// v123 is the value that the 123 joker takes in the formation.
// vLeader is the value that leader joker takes."
func evalFormationT0(troops []*cards.Troop) (formation *cards.Formation, v123 int, vLeader int) {
	v123 = 0    //always zero
	vLeader = 0 //always zero
	value, color := evalFormationValueColor(troops)
	if color {
		formation = &cards.FBattalion
	}
	line := evalFormationLine(troops)
	if formation != nil { // battalion or wedge
		if line {
			formation = &cards.FWedge
		}
	} else { //no battalion or wedge
		if value {
			formation = &cards.FPhalanx
		}
		if formation == nil { // no battalion, wedge or phalanx
			if line {
				formation = &cards.FSkirmish
			} else { // no battalion wedge  phalanx Line
				formation = &cards.FHost
			}
		}
	}
	return formation, v123, vLeader
}

//evalFormationValueColor checks for same value and color.
func evalFormationValueColor(troops []*cards.Troop) (value, color bool) {
	value = true
	color = true
	for i, v := range troops {
		if i != len(troops)-1 {
			if value && v.Value() != troops[i+1].Value() {
				value = false
				if !color {
					break
				}
			}
			if color && v.Color() != troops[i+1].Color() {
				color = false
				if !value {
					break
				}
			}
		}
	}
	return value, color
}

// evalFormationLine evalute a line assums the cards are sorted bigest first.
func evalFormationLine(troops []*cards.Troop) (line bool) {
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

// evalFormationT1Line evaluate a line with one joker.
// Expect troops to sorted biggest first and trimmed no nil values
// The skipValue is the jokers strenght if zero it is not determent,
func evalFormationT1Line(troops []*cards.Troop) (line bool, skipValue int) {
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
					player.Strenght >= opPlayer.Strenght {
					ok = true
				} else {
					ok = false
					eks = make([]int, len(opPlayer.Troops))
					copy(eks, opPlayer.Troops[:])
				}
			} else { // opponent no formation
				sortInt(opPlayer.Troops[:]) //============Sort Oponnent Troops=====================
				opTroops := make([]*cards.Troop, 0, 3)
				for _, v := range opPlayer.Troops {
					opTroop, ok := cards.DrTroop(v)
					if ok {
						opTroops = append(opTroops, opTroop)
					}
				}
				if len(opTroops) > 1 {
					ok = claimFlagOppentPlayedCard(player.Formation, player.Strenght, opTroops, mud)
				}

				if !ok {
					ok, eks = claimFlagSimulation(player.Formation, player.Strenght, opPlayer.Troops, mud, fog, unPlayCards)
				}
			}
		}
	}
	if ok {
		player.Won = true
	}
	return ok, eks
}

//claimFlagSimulation simulate all formation to check for wining formation.
//that would falsify the claim.
//ok true if no formation exist.
func claimFlagSimulation(
	formation *cards.Formation,
	strenght int,
	oppTroopixs [4]int,
	isMud, isFog bool,
	unPlayCards []int) (ok bool, eks []int) {

	playedCards := played(oppTroopixs[:])
	opCards := make([]int, 4) // copy
	copy(opCards, oppTroopixs[:])
	simNoCards := 3 - playedCards
	simPlayedCards := 3
	if isMud {
		simNoCards = simNoCards + 1
		simPlayedCards = 4
	}
	simCards := make([]int, 4) //copy
	switch simNoCards {
	case 1:
		ok = true
		for _, card := range unPlayCards {
			copy(simCards, opCards)
			simCards[len(opCards)-1] = card
			simFormation, simStrenght := evalSim(isMud, isFog, simCards, simPlayedCards)
			if simFormation.Value > formation.Value ||
				(simFormation.Value == formation.Value && simStrenght > strenght) {
				eks = simCards
				ok = false
				break
			}
		}
	case 2:
		match := math.Perm2(len(unPlayCards), func(v [2]int) bool {
			copy(simCards, opCards)
			simCards[len(opCards)-1] = unPlayCards[v[1]]
			simCards[len(opCards)-2] = unPlayCards[v[0]]
			simFormation, simStrenght := evalSim(isMud, isFog, simCards, simPlayedCards)
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
	case 3:
		match := math.Perm3(len(unPlayCards), func(v [3]int) bool {
			copy(simCards, opCards)
			simCards[len(opCards)-1] = unPlayCards[v[2]]
			simCards[len(opCards)-2] = unPlayCards[v[1]]
			simCards[len(opCards)-3] = unPlayCards[v[0]]
			simFormation, simStrenght := evalSim(isMud, isFog, simCards, simPlayedCards)
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
	case 4:
		match := math.Perm4(len(unPlayCards), func(v [4]int) bool {
			copy(simCards, opCards)
			simCards[len(opCards)-1] = unPlayCards[v[3]]
			simCards[len(opCards)-2] = unPlayCards[v[2]]
			simCards[len(opCards)-3] = unPlayCards[v[1]]
			simCards[len(opCards)-4] = unPlayCards[v[0]]
			simFormation, simStrenght := evalSim(isMud, isFog, simCards, simPlayedCards)
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

	return ok, eks
}

//claimFlagOppentPlayedCard checks if a flag can be claimed base on caclulated max formation.
//These calculation need two troop to be played and does no include
//Host just a sum.
func claimFlagOppentPlayedCard(
	formation *cards.Formation,
	strenght int,
	opTroops []*cards.Troop,
	isMud bool) (ok bool) {

	value, color := evalFormationValueColor(opTroops)
	line := true
	if !isMud {
		line, _ = evalFormationT1Line(opTroops) // could be better but for now ok. must be sorted
	}
	wedge := line && color

	switch formation {
	case &cards.FWedge:
		if !wedge {
			ok = true
		}
	case &cards.FPhalanx:
		if !wedge && !value {
			ok = true
		} else {
			m := 3
			if isMud {
				m = 4
			}
			if value && opTroops[0].Value()*m < strenght {
				ok = true
			}
		}
	case &cards.FBattalion:
		if !wedge && !value && !color {
			ok = true
		}
	case &cards.FSkirmish:
		if !color && !value && !line {
			ok = true
		}
	}
	return ok
}

//UsedTac collects the used tactic cards.
func (flag *Flag) UsedTac() (v [2][]int) {
	v[0] = usedTacTac(flag.Players[0].Env[:], flag.Players[0].Troops[:])
	v[1] = usedTacTac(flag.Players[1].Env[:], flag.Players[1].Troops[:])
	return v
}

//UsedTacTac collects the used tactic cards.
func usedTacTac(Env []int, troops []int) (tacs []int) {
	tacs = make([]int, 0, 5)
	for _, e := range Env {
		if e != 0 {
			tacs = append(tacs, e)
		}
	}
	for _, cardix := range troops {
		if cardix != 0 {
			if cards.IsTac(cardix) {
				tacs = append(tacs, cardix)
			}
		}
	}
	return tacs
}

//Free check if there is space for a card on the flag.
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

//Env makes a slice with envirement tactic cards.
//may be empty.
func (flag *Flag) Env(pix int) (env []int) {
	env = removeZeros(flag.Players[pix].Env[:])
	return env
}
func removeZeros(list []int) (clean []int) {
	clean = make([]int, 0, len(list))
	for _, v := range list {
		if v != 0 {
			clean = append(clean, v)
		}
	}
	return clean
}

//Troops makes a slice with troops maybe empty.
func (flag *Flag) Troops(pix int) (troops []int) {
	troops = removeZeros(flag.Players[pix].Troops[:])
	return troops
}

//Claimed return true if the flag is claimed.
func (flag *Flag) Claimed() bool {
	return flag.Players[0].Won || flag.Players[1].Won
}

//Won returns players won.
func (flag *Flag) Won() (res [2]bool) {
	res[0] = flag.Players[0].Won
	res[1] = flag.Players[1].Won
	return res
}

//Formations returns players formations.
func (flag *Flag) Formations() (form [2]*cards.Formation) {
	form[0] = flag.Players[0].Formation
	form[1] = flag.Players[1].Formation
	return form
}
