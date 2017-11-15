package game

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
	math "github.com/rezder/go-math/int"
)

// Flag a flag both players cards connected to a cone.
type Flag struct {
	Players   [2]FlagPlayer
	Positions [2]pos.Card
	ConePos   pos.Cone
	IsWon     bool
	IsMud     bool
	IsFog     bool
}

func (f *Flag) String() string {
	return fmt.Sprintf("Flag{Players:%v,Positions:%v,ConePos:%v,IsWon:%v,IsMud:%v,IsFog:%v}",
		f.Players, f.Positions, f.ConePos, f.IsWon, f.IsMud, f.IsFog)
}
func (f *Flag) Copy() (c *Flag) {
	if f != nil {
		c = new(Flag)
		c.ConePos = f.ConePos
		c.IsWon = f.IsWon
		c.IsMud = f.IsMud
		c.IsFog = f.IsFog
		for i, player := range f.Players {
			c.Players[i] = player.Copy()
		}
	}
	return c
}

// FlagPlayer the player in a flag.
type FlagPlayer struct {
	Troops  []card.Troop
	Morales []card.Morale
	Envs    []card.Env
}

func (f FlagPlayer) Copy() (c FlagPlayer) {
	if f.Troops != nil {
		c.Troops = make([]card.Troop, len(f.Troops))
		copy(c.Troops, f.Troops)
	}
	if f.Morales != nil {
		c.Morales = make([]card.Morale, len(f.Morales))
		copy(c.Morales, f.Morales)
	}
	if f.Envs != nil {
		c.Envs = make([]card.Env, len(f.Envs))
		copy(c.Envs, f.Envs)
	}
	return c
}

// NewFlag creates a flag.
func NewFlag(ix int, posCards PosCards, conePos [10]pos.Cone) (f *Flag) {
	f = new(Flag)
	f.ConePos = conePos[ix+1]
	f.IsWon = f.ConePos.IsWon()
	f.Positions[0] = pos.CardAll.Players[0].Flags[ix]
	f.Positions[1] = pos.CardAll.Players[1].Flags[ix]
	if !f.IsWon {
		cards := posCards.SortedCards(f.Positions[0])
		f.Players[0].Troops = cards.Troops
		f.Players[0].Morales = cards.Morales
		f.Players[0].Envs = cards.Envs
		cards = posCards.SortedCards(f.Positions[1])
		f.Players[1].Troops = cards.Troops
		f.Players[1].Morales = cards.Morales
		f.Players[1].Envs = cards.Envs
		envs := make([]card.Env, len(f.Players[0].Envs)+len(f.Players[1].Envs))
		if len(envs) > 0 {
			copy(envs, f.Players[0].Envs)
			copy(envs[len(f.Players[0].Envs):], f.Players[1].Envs)
			for _, env := range envs {
				if env.IsMud() {
					f.IsMud = true
				}
				if env.IsFog() {
					f.IsFog = true
				}
			}
		}
	}
	return f
}

//IsTroopPlayable returns true if it is possible to play a troop
//on the flag by the player.
func (f *Flag) IsTroopPlayable(player int) (isPlayable bool) {
	if f.IsWon {
		isPlayable = false
	} else {
		no := len(f.Players[player].Troops) + len(f.Players[player].Morales)
		if no < 3 || (f.IsMud && no < 4) {
			isPlayable = true
		}
	}
	return isPlayable
}

//IsMoralePlayable returns true if it is possible to play a morale
// tactic card on the flag by the player.
func (f *Flag) IsMoralePlayable(player int) bool {
	return f.IsTroopPlayable(player)
}

//IsEnvPlayable returns true if it is possible to play a enviroment
// tactic card on the flag by the player.
func (f *Flag) IsEnvPlayable(player int) bool {
	return !f.IsWon
}

//HasFormation returns true if the player have
//enough cards to make a formation.
func (f *Flag) HasFormation(player int) bool {
	if !f.IsWon {
		return !f.IsTroopPlayable(player)
	}
	return false
}

// FormationSize the size of formation 3 or 4.
func (f *Flag) FormationSize() (size int) {
	return formationSize(f.IsMud)
}

//Formation calculates the formation and strenght.
func (f *Flag) Formation(playix int) (formation *card.Formation, strenght int) {
	formation, strenght = eval(f.Players[playix].Troops, f.Players[playix].Morales, f.IsMud, f.IsFog)
	return formation, strenght
}
func (f *Flag) PlayerFormationSize(playerix int) int {
	return len(f.Players[playerix].Troops) + len(f.Players[playerix].Morales)
}

//IsClaimable return true it the player can succesfully claim the flag.
func (f *Flag) IsClaimable(player int, deckTroops []card.Troop) (isClaim bool, exCardixs []card.Card) {
	if f.HasFormation(player) {
		formation, strenght := eval(f.Players[player].Troops, f.Players[player].Morales, f.IsMud, f.IsFog)
		if formation == &card.FWedge && (strenght == 27 && !f.IsMud || strenght == 34 && f.IsMud) {
			isClaim = true
		} else {
			var exTroops []card.Troop
			opponent := opp(player)
			if f.HasFormation(opponent) {
				oppFormation, oppStrenght := eval(f.Players[opponent].Troops, f.Players[opponent].Morales, f.IsMud, f.IsFog)
				if oppFormation.Value < formation.Value ||
					(oppFormation.Value == formation.Value && oppStrenght <= strenght) {
					isClaim = true
				} else {
					exTroops = f.Players[opponent].Troops
				}
			} else {
				if len(f.Players[opponent].Troops) > 1 {
					oppFormation, oppStrenght := estimateFormation(f.IsFog, f.IsMud,
						f.Players[opponent].Troops, f.Players[opponent].Morales)
					if oppFormation.Value < formation.Value ||
						(oppFormation.Value == formation.Value && oppStrenght <= strenght) {
						isClaim = true
					}
				}
				if !isClaim {
					isClaim, exTroops = isFlagClaimableSim(formation, strenght, f.IsFog, f.IsMud,
						f.Players[opponent].Troops, f.Players[opponent].Morales, deckTroops)
				}
			}
			if !isClaim {
				exCardixs = make([]card.Card, 0, 4)
				for _, troop := range exTroops {
					exCardixs = append(exCardixs, card.Card(troop))
				}
				for _, morale := range f.Players[opponent].Morales {
					exCardixs = append(exCardixs, card.Card(morale))
				}
			}
		}
	}
	return isClaim, exCardixs
}
func formationSize(isMud bool) int {
	if isMud {
		return 4
	}
	return 3
}
func estimateFormation(
	isFog, isMud bool,
	troops []card.Troop,
	morales []card.Morale) (form *card.Formation, strenght int) {
	formSize := formationSize(isMud)
	noMissing := formSize - len(troops)
	if isFog {
		form, strenght = estimateFormationFog(troops, noMissing)
	} else {
		if len(troops) < 2 {
			form, strenght = topFormation(formSize)
		} else {
			isUniqStrenght, isUnigColor := evalUniqueStrenghtColor(troops)
			for _, troop := range troops {
				strenght = strenght + troop.Strenght()
			}
			isLine := false
			missLineStrenght := strenght
			if !isUniqStrenght {
				isLine, missLineStrenght = estimateLine(troops, noMissing, strenght)
			}
			if isUnigColor {
				if isLine {
					form = &card.FWedge
					strenght = strenght + missLineStrenght
				} else {
					form = &card.FBattalion
					strenght = strenght + 10*noMissing
				}
			} else if isUniqStrenght {
				form = &card.FPhalanx
				strenght = strenght + noMissing*troops[0].Strenght()
			} else if isLine {
				form = &card.FSkirmish
				strenght = strenght + missLineStrenght
			} else {
				form, strenght = estimateFormationFog(troops, noMissing)
			}
		}
	}
	return form, strenght
}
func topFormation(formSize int) (form *card.Formation, strenght int) {
	form = &card.FWedge
	strenght = 0
	for i := 0; i < formSize; i++ {
		strenght = strenght + 10 - i
	}
	return form, strenght
}
func estimateLine(troops []card.Troop, noMissing int, troopsStrenght int) (isLine bool, missStrenght int) {
	isLine = true
	steps := noMissing
	for i, troop := range troops {
		if i != len(troops)-1 {
			diffStrength := troop.Strenght() - troops[i+1].Strenght()
			if diffStrength > steps+1 || diffStrength == 0 {
				isLine = false
				break
			} else if diffStrength != 1 {
				steps = steps - diffStrength + 1
			}
		}
	}
	if isLine {
		for i := 0; i < noMissing; i++ {
			missStrenght = missStrenght + 10 - i
		}
	}
	return isLine, missStrenght
}
func estimateFormationFog(
	troops []card.Troop,
	noMissing int) (form *card.Formation, strenght int) {
	form = &card.FHost
	strenght = noMissing * 10
	for _, troop := range troops {
		strenght = strenght + troop.Strenght()
	}
	return form, strenght
}

func isFlagClaimableSim(
	targetForm *card.Formation,
	targetStrenght int,
	isFog, isMud bool,
	troops []card.Troop,
	morales []card.Morale,
	deckTroops []card.Troop) (isClaim bool, exTroops []card.Troop) {
	formSize := formationSize(isMud)
	noMissing := formSize - len(troops) - len(morales)
	switch noMissing {
	case 1:
		isClaim, exTroops = sim1Card(targetForm, targetStrenght, deckTroops, troops, morales, isMud, isFog)
	case 2:
		isClaim, exTroops = sim2Cards(targetForm, targetStrenght, deckTroops, troops, morales, isMud, isFog)
	case 3:
		isClaim, exTroops = sim3Cards(targetForm, targetStrenght, deckTroops, troops, morales, isMud, isFog)
	case 4:
		isClaim, exTroops = sim4Cards(targetForm, targetStrenght, deckTroops, troops, morales, isMud, isFog)
	default:
		panic(fmt.Sprintf("Only up till 4 cards has been implemented no cards: %v", noMissing))

	}
	return isClaim, exTroops
}
func sim1Card(
	targetFormation *card.Formation,
	targetStrenght int,
	deckTroops, troops []card.Troop,
	morales []card.Morale,
	isMud, isFog bool) (isClaim bool, exTroops []card.Troop) {
	noMissing := 1
	isClaim = true
	for _, simTroop := range deckTroops {
		simTroops := make([]card.Troop, len(troops), len(troops)+noMissing)
		copy(simTroops, troops)
		simTroops = simTroop.AppendStrSorted(simTroops)
		simFormation, simStrenght := eval(simTroops, morales, isMud, isFog)
		if simFormation.Value > targetFormation.Value ||
			(simFormation.Value == targetFormation.Value && simStrenght > targetStrenght) {
			exTroops = simTroops
			isClaim = false
			break
		}
	}
	return isClaim, exTroops
}
func sim2Cards(
	targetFormation *card.Formation,
	targetStrenght int,
	deckTroops, troops []card.Troop,
	morales []card.Morale,
	isMud, isFog bool) (isClaim bool, exTroops []card.Troop) {
	noMissing := 2
	isClaim = true
	math.Perm2(len(deckTroops), func(v [2]int) bool {
		simTroops := make([]card.Troop, len(troops), len(troops)+noMissing)
		copy(simTroops, troops)
		simTroops = deckTroops[v[0]].AppendStrSorted(simTroops)
		simTroops = deckTroops[v[1]].AppendStrSorted(simTroops)
		simFormation, simStrenght := eval(simTroops, morales, isMud, isFog)
		if simFormation.Value > targetFormation.Value ||
			(simFormation.Value == targetFormation.Value && simStrenght > targetStrenght) {
			exTroops = simTroops
			isClaim = false
			return true //stop loop
		}
		return false //continue loop
	})
	return isClaim, exTroops
}
func sim3Cards(
	targetFormation *card.Formation,
	targetStrenght int,
	deckTroops, troops []card.Troop,
	morales []card.Morale,
	isMud, isFog bool) (isClaim bool, exTroops []card.Troop) {

	noMissing := 3
	isClaim = true
	math.Perm3(len(deckTroops), func(v [3]int) bool {
		simTroops := make([]card.Troop, len(troops), len(troops)+noMissing)
		copy(simTroops, troops)
		simTroops = deckTroops[v[0]].AppendStrSorted(simTroops)
		simTroops = deckTroops[v[1]].AppendStrSorted(simTroops)
		simTroops = deckTroops[v[2]].AppendStrSorted(simTroops)
		simFormation, simStrenght := eval(simTroops, morales, isMud, isFog)
		if simFormation.Value > targetFormation.Value ||
			(simFormation.Value == targetFormation.Value && simStrenght > targetStrenght) {
			exTroops = simTroops
			isClaim = false
			return true //stop loop
		}
		return false //continue loop
	})
	return isClaim, exTroops
}
func sim4Cards(
	targetFormation *card.Formation,
	targetStrenght int,
	deckTroops, troops []card.Troop,
	morales []card.Morale,
	isMud, isFog bool) (isClaim bool, exTroops []card.Troop) {
	noMissing := 4
	isClaim = true
	math.Perm4(len(deckTroops), func(v [4]int) bool {
		simTroops := make([]card.Troop, 0, noMissing)
		simTroops = deckTroops[v[0]].AppendStrSorted(simTroops)
		simTroops = deckTroops[v[1]].AppendStrSorted(simTroops)
		simTroops = deckTroops[v[2]].AppendStrSorted(simTroops)
		simTroops = deckTroops[v[3]].AppendStrSorted(simTroops)
		simFormation, simStrenght := eval(simTroops, morales, isMud, isFog)
		if simFormation.Value > targetFormation.Value ||
			(simFormation.Value == targetFormation.Value && simStrenght > targetStrenght) {
			exTroops = simTroops
			isClaim = false
			return true //stop loop
		}
		return false //continue loop
	})
	return isClaim, exTroops
}
func maxFormation(
	aFormation *card.Formation,
	aStrenght int,
	bFormation *card.Formation,
	bStrenght int) (*card.Formation, int) {
	if aFormation.Value > bFormation.Value ||
		(aFormation.Value == bFormation.Value && aStrenght > bStrenght) {
		return aFormation, aStrenght
	}
	return bFormation, bStrenght
}

//evalMudExces evaluate the special case where mud has been removed and left
//one card in excess.
func evalMudExcess(
	troops []card.Troop,
	morales []card.Morale,
	ismud, isfog bool) (formation *card.Formation, strenght int) {
	noTroops := len(troops)
	formation = &card.FHost
	for i := 0; i < 4; i++ {
		simTroops := make([]card.Troop, 0, 4)
		simMorales := make([]card.Morale, 0, 4)
		for j := 0; j < 4; j++ {
			if i != j {
				if j < noTroops {
					simTroops = append(simTroops, troops[j])
				} else {
					simMorales = append(simMorales, morales[j-noTroops])
				}
			}
		}
		nextForm, nextStrenght := eval(simTroops, simMorales, ismud, isfog)
		formation, strenght = maxFormation(formation, strenght, nextForm, nextStrenght)
	}
	return formation, strenght
}

func eval(
	troops []card.Troop,
	morales []card.Morale,
	isMud, isFog bool) (formation *card.Formation, strenght int) {

	if len(troops)+len(morales) == formationSize(isMud)+1 {
		formation, strenght = evalMudExcess(troops, morales, isMud, isFog)
	} else {
		if isFog {
			formation = &card.FHost
			for _, troop := range troops {
				strenght = strenght + troop.Strenght()
			}
			for _, morale := range morales {
				strenght = strenght + morale.MaxStrenght()
			}
		} else {
			var jokerStrenghts []int
			formation, jokerStrenghts = evalFormation(troops, morales)
			for _, troop := range troops {
				strenght = strenght + troop.Strenght()
			}
			for _, joker := range jokerStrenghts {
				strenght = strenght + joker
			}
		}
	}
	return formation, strenght
}
func evalFormation(
	troops []card.Troop,
	morales []card.Morale) (formation *card.Formation, jokerStrenghts []int) {
	switch len(morales) {
	case 0:
		formation = evalFormationTroops(troops)
	case 1:
		formation, jokerStrenghts = evalFormationTac(troops, morales[0])
	case 2:
		formation, jokerStrenghts = evalFormationTacs(troops, morales)
	case 3:
		formation = &card.FBattalion
		for _, morale := range morales {
			jokerStrenghts = append(jokerStrenghts, morale.MaxStrenght())
		}
	}
	return formation, jokerStrenghts
}
func evalFormationTroops(
	troops []card.Troop) (formation *card.Formation) {
	isUniqStrenght, isUnigColor := evalUniqueStrenghtColor(troops)
	isLine := evalLine(troops)
	if isUnigColor {
		if isLine {
			formation = &card.FWedge
		} else {
			formation = &card.FBattalion
		}
	} else {
		if isUniqStrenght {
			formation = &card.FPhalanx
		} else {
			if isLine {
				formation = &card.FSkirmish
			} else {
				formation = &card.FHost
			}
		}
	}
	return formation
}

//func evalLine
func evalLine(troops []card.Troop) (isLine bool) {
	isLine = true
	for i, troop := range troops {
		if i != len(troops)-1 {
			if troop.Strenght() != troops[i+1].Strenght()+1 {
				isLine = false
				break
			}
		}
	}
	return isLine
}

//evalFormationTacs evaluate a formation with two jokers.
// troops must be sorted biggest first.
func evalFormationTacs(
	troops []card.Troop,
	morales []card.Morale) (formation *card.Formation, jokerStrenghts []int) {
	_, isUnigColor := evalUniqueStrenghtColor(troops)
	if !morales[0].IsLeader() && !morales[1].IsLeader() {
		if isUnigColor {
			formation = &card.FBattalion
		} else {
			formation = &card.FHost
		}
		for _, morale := range morales {
			jokerStrenghts = append(jokerStrenghts, morale.MaxStrenght())
		}
	} else if morales[0].HasStrenght() || morales[1].HasStrenght() {
		formation, jokerStrenghts = evalLeader8(troops)
	} else {
		formation, jokerStrenghts = evalLeader123(troops)
	}
	return formation, jokerStrenghts
}

//evalLeader8Troop find the formation and the leader value for a flag
//with a leader,8 morale tactic card and one troop.
func evalLeader8Troop(troop card.Troop) (formation *card.Formation, vLeader int) {
	if troop.Strenght() == 10 {
		formation = &card.FWedge
		vLeader = 9
	} else if troop.Strenght() == 9 {
		formation = &card.FWedge
		vLeader = 10
	} else if troop.Strenght() == 8 {
		formation = &card.FPhalanx
		vLeader = 8
	} else if troop.Strenght() == 7 {
		formation = &card.FWedge
		vLeader = 9
	} else if troop.Strenght() == 6 {
		formation = &card.FWedge
		vLeader = 7
	} else {
		formation = &card.FBattalion
		vLeader = 10
	}
	return formation, vLeader
}
func evalLeader8Line(troops []card.Troop) (vLeader int) {
	if troops[0].Strenght() == 10 && troops[1].Strenght() == 9 {
		vLeader = 7
	} else if troops[0].Strenght() == 10 && troops[1].Strenght() == 7 {
		vLeader = 9
	} else if troops[0].Strenght() == 9 && troops[1].Strenght() == 7 {
		vLeader = 10
	} else if troops[0].Strenght() == 9 && troops[1].Strenght() == 6 {
		vLeader = 7
	} else if troops[0].Strenght() == 7 && troops[1].Strenght() == 6 {
		vLeader = 9
	} else if troops[0].Strenght() == 7 && troops[1].Strenght() == 5 {
		vLeader = 6
	} else if troops[0].Strenght() == 6 && troops[1].Strenght() == 5 {
		vLeader = 7
	}
	return vLeader
}

//evalLeader8 find the formation and the leader value for a flag
//with a leader and the 8 morale tactic card.
//troops must be sorted biggest first.
func evalLeader8(troops []card.Troop) (formation *card.Formation, jokerStrenghts []int) {
	_, isUniqColor := evalUniqueStrenghtColor(troops)
	vLeader := 0
	if len(troops) == 1 {
		formation, vLeader = evalLeader8Troop(troops[0])
	} else { // Two troops
		vLeader = evalLeader8Line(troops)
		if troops[0].Strenght() == 8 && troops[1].Strenght() == 8 {
			formation = &card.FPhalanx
			vLeader = 8
		} else {
			if isUniqColor {
				if vLeader != 0 {
					formation = &card.FWedge
				} else {
					formation = &card.FBattalion
					vLeader = 10
				}
			} else {
				if vLeader != 0 {
					formation = &card.FSkirmish
				} else {
					formation = &card.FHost
					vLeader = 10
				}
			}
		}
	}
	return formation, []int{vLeader, 8}
}

//evalLeader123Troop finds a formation and the morale values for a flag
//with a leader,123 morale tactic card and one troop.
func evalLeader123Troop(troop card.Troop) (formation *card.Formation, v123, vLeader int) {
	if troop.Strenght() < 6 {
		formation = &card.FWedge
		if troop.Strenght() == 1 {
			v123 = 2
			vLeader = 3
		} else if troop.Strenght() == 2 {
			v123 = 3
			vLeader = 4
		} else if troop.Strenght() == 3 {
			v123 = 2
			vLeader = 4
		} else if troop.Strenght() == 4 {
			v123 = 3
			vLeader = 5
		} else if troop.Strenght() == 5 {
			v123 = 3
			vLeader = 4
		}
	} else {
		formation = &card.FBattalion
	}
	return formation, v123, vLeader
}

func evalLeader123Line(troops []card.Troop) (v123, vLeader int) {
	if troops[0].Strenght() == 3 && troops[1].Strenght() == 2 {
		v123 = 1
		vLeader = 4
	} else if troops[0].Strenght() == 4 && troops[1].Strenght() == 1 {
		v123 = 2
		vLeader = 3
	} else if troops[0].Strenght() == 3 && troops[1].Strenght() == 1 {
		v123 = 2
		vLeader = 4
	} else if troops[0].Strenght() == 5 && troops[1].Strenght() == 3 {
		v123 = 2
		vLeader = 4
	} else if troops[0].Strenght() == 4 && troops[1].Strenght() == 3 {
		v123 = 2
		vLeader = 5
	} else if troops[0].Strenght() == 2 && troops[1].Strenght() == 1 {
		v123 = 3
		vLeader = 4
	} else if troops[0].Strenght() == 6 && troops[1].Strenght() == 5 {
		v123 = 3
		vLeader = 4
	} else if troops[0].Strenght() == 4 && troops[1].Strenght() == 2 {
		v123 = 3
		vLeader = 5
	} else if troops[0].Strenght() == 6 && troops[1].Strenght() == 4 {
		v123 = 3
		vLeader = 5
	} else if troops[0].Strenght() == 5 && troops[1].Strenght() == 4 {
		v123 = 3
		vLeader = 6
	}
	return v123, vLeader
}

//evalLeader123 finds a formation and the leader value for a flag
//with a leader and the 123 morale tactic card.
//troops must be sorted biggest first.
//jokerStrenghts is the values the joker takes if used.
func evalLeader123(troops []card.Troop) (formation *card.Formation, jokerStrenghts []int) {
	v123 := 0
	vLeader := 0
	if len(troops) == 1 {
		formation, v123, vLeader = evalLeader123Troop(troops[0])
	} else { // two troops
		isUniqStrenght, isUniqColor := evalUniqueStrenghtColor(troops)
		if isUniqStrenght && troops[0].Strenght() < 4 {
			vLeader = troops[0].Strenght()
			v123 = troops[0].Strenght()
			formation = &card.FPhalanx
		} else {
			v123, vLeader = evalLeader123Line(troops)
		}
		if formation == nil { //no phalanx
			if isUniqColor {
				if v123 != 0 { //line
					formation = &card.FWedge
				} else {
					formation = &card.FBattalion
					v123 = 3
					vLeader = 10
				}
			} else {
				if v123 != 0 { //line
					formation = &card.FSkirmish
				} else {
					formation = &card.FHost
					v123 = 3
					vLeader = 10
				}
			}
		}
	}
	return formation, []int{v123, vLeader}
}
func evalFormationTac(
	troops []card.Troop,
	morale card.Morale) (formation *card.Formation, jokerStrenghts []int) {
	isUniqStrenght, isUniqColor := evalUniqueStrenghtColor(troops)
	isLine, jokerStrenght := evalLineMorale(troops, morale)
	if isUniqColor {
		if isLine {
			formation = &card.FWedge
		} else {
			formation = &card.FBattalion
			jokerStrenght = morale.MaxStrenght()
		}
	} else {
		if isUniqStrenght && morale.ValidStrenght(troops[0].Strenght()) {
			formation = &card.FPhalanx
			jokerStrenght = troops[0].Strenght()
		} else if isLine {
			formation = &card.FSkirmish
		} else {
			formation = &card.FHost
			jokerStrenght = morale.MaxStrenght()
		}
	}
	return formation, []int{jokerStrenght}
}

// evalLineMorale evaluate a line with one joker.
// Expect troops to sorted biggest first.
func evalLineMorale(troops []card.Troop, morale card.Morale) (isLine bool, jokerStrenght int) {
	isLine = true
	skipStrenght := 0
	for i, troop := range troops {
		if i != len(troops)-1 {
			if troop.Strenght() != troops[i+1].Strenght()+1 {
				if troop.Strenght() == troops[i+1].Strenght()+2 && skipStrenght == 0 {
					skipStrenght = troop.Strenght() - 1
				} else {
					isLine = false
					break
				}
			}
		}
	}
	if skipStrenght == 0 && isLine {
		top := troops[0].Strenght() + 1
		butt := troops[len(troops)-1].Strenght() - 1
		if morale.ValidStrenght(top) {
			jokerStrenght = top
		} else if morale.ValidStrenght(butt) {
			jokerStrenght = butt
		} else {
			isLine = false
		}
	} else if skipStrenght > 0 {
		if morale.ValidStrenght(skipStrenght) {
			jokerStrenght = skipStrenght
		} else {
			isLine = false
		}
	}
	return isLine, jokerStrenght
}
func evalUniqueStrenghtColor(troops []card.Troop) (isStrenght, isColor bool) {
	isStrenght = true
	isColor = true
	for i, troop := range troops {
		if i != len(troops)-1 {
			if isStrenght && troop.Strenght() != troops[i+1].Strenght() {
				isStrenght = false
				if !isColor {
					break
				}
			}
			if isColor && troop.Color() != troops[i+1].Color() {
				isColor = false
				if !isStrenght {
					break
				}
			}
		}
	}
	return isStrenght, isColor
}
