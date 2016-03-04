package flag

import (
	"errors"
	"rezder.com/game/card/battleline/cards"
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

func (flag *Flag) Remove(cardix int, playerix int) (err error) {
	//TODO remove a card
	return err
}
func (flag *Flag) Set(cardix int, playerix int) (err error) {
	card, err := cards.DrCard(cardix)
	if err == nil {
		switch tcard := (*card).(type) {
		case cards.Tactic:
			if tcard.Type() == cards.T_Env {
				place := set_card(cardix, flag.players[playerix].env[:])
				if place != -1 {
					flag.updateFormations()
				} else {
					err = errors.New("No available slots")
				}
			} else if tcard.Type() == cards.T_Morale {
				flag.set_Troop(cardix, playerix)
			} else {
				err = errors.New("Illegal tactic card")
			}
		case cards.Troop:
			flag.set_Troop(cardix, playerix)
		default:
			panic("No supprted type")
		}
	} else {
		panic("Card do not exist")
	}
	return err
}
func (flag *Flag) updateFormations() {
	flag.updateFormation(0)
	flag.updateFormation(1)
}
func (flag *Flag) updateFormation(playerix int) {
	player := &flag.players[playerix]
	troops := player.troops[:]
	envs := player.env[:]
	sort.Ints(troops) //increaing order
	sort.Ints(envs)
	//ix:=sort.Search(len(envs), f func(i int) bool{
	//	return envs[i]==cards.TC_Mud
	//})
	var mud bool
	var fog bool
	for _, card := range envs {
		if card == cards.TC_Mud {
			mud = true
		} else if card == cards.TC_Fog {
			fog = true
		}
	}
	played := 0
	for i, troop := range troops {
		if troop != 0 {
			played = 4 - i
			break
		}
	}
	if mud && played == 4 {
		if fog {
			player.formation = &cards.F_Host
			player.strenght = strenght(troops, vm123, vmLeader)
		}
		formation, v123, vLeader := evalFormation(troops)
		player.formation = &formation
		player.strenght = strenght(troops, v123, vLeader)
	} else if played == 3 {
		if fog {
			player.formation = &cards.F_Host
			player.strenght = strenght(troops, vm123, vmLeader)
		} else {
			formation, v123, vLeader := evalFormation(troops)
			player.formation = &formation
			player.strenght = strenght(troops, v123, vLeader)
		}
	} else {
		player.formation = nil
		player.strenght = 0
	}

}

//evalFormation evaluate the formation for a full formation 3 or 4 and no fog.
//Plan split the problem in case of tactic cards.
//Case 1: No tactic cards
//Case 2: One tactic card
//Case 3: Two tactic cards
//Case 4: Three tactic cards
//3 or 4 cards should not make a difference
func evalFormation(troops []int) (formation cards.Formation, v123 int, vLeader int) {
	//TODO
	return formation, v123, vLeader
}

//claimFlag Claims a flag if possible.
//Its only possible to claim a flag if player have a formation, if
//both players have a formation it is easy, or if player have the highest formation.
//The main task is calculate the highest possible formation.
//I do not have a plan yet but I think we need brute force, but
//If we are missing four cards. Its 57*56*55*54/24 permutain and how do I generate them.
//The player's formation do limit the possible formation to the ones that are better but again
//that may be many all to be exact, but if we find one that is better we can stop.
// I see 2 option:
//      1) Just try all permution and if one is higher stop.
//      2) For every combination try to make one with the remaining cards
//
func (flag *Flag) claimFlag(playerix int, unPlayCards []int) (ok bool) {
	//TODO
	player := &flag.players[playerix]
	return ok
}
func strenght(form []int, v123 int, vLeader int) (st int) {
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
			panic("Card do not exist or is not legal")
		}
	}
	return st
}
func set_card(cardix int, v []int) (place int) {
	place = -1
	for i, c := range v {
		if c != 0 {
			v[i] = cardix
			place = i
			break
		}
	}
	return place
}
func (flag *Flag) set_Troop(cardix int, playerix int) (err error) {
	place := set_card(cardix, flag.players[playerix].troops[:])
	if place != -1 {
		flag.updateFormation(playerix)
	} else {
		err = errors.New("No available slots")
	}
	return err
}
