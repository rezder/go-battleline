package combi

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/game/card"
)

const (
	tbNone     = 0
	tbRank     = 1
	tbStrenght = 2
)

var (
	combinations3, combinations4 []*Combination
	tbNames                      [3]string
)

func init() {
	combinations3 = createCombi(3)
	combinations4 = createCombi(4)
	tbNames = [3]string{"None", "Rank", "Strenght"}
}

//Combinations returns the all the possible combinations.
func Combinations(size int) []*Combination {
	if size == 4 {
		return combinations4
	}
	return combinations3
}

//CombinationsMud returns the all the possible combinations.
func CombinationsMud(isMud bool) []*Combination {
	if isMud {
		return combinations4
	}
	return combinations3
}

// RankHost returns host rank
func RankHost(size int) int {
	return len(Combinations(size))
}

// RankTieBreaker returns the tiebreaker rule.
func RankTieBreaker(rank, size int) TieBreaker {
	return Combinations(size)[rank-1].TieBreaker
}

//Combination a battleline formation and strength
type Combination struct {
	Rank      int
	Formation card.Formation
	Strength  int
	//Troops is all the cards that can be used to create the formation.
	//per color
	Troops map[int][]card.Troop
	TieBreaker
}

func (c *Combination) String() string {
	if c == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{Rank:%v Strength:%v Formation:%v Troops:%v TieBreaker:%v",
		c.Rank, c.Strength, c.Formation, c.Troops, c.TieBreaker)
}

//createCombi create all the possible combinations for the specified number of
//cards 3 or 4.
func createCombi(cardsNo int) (combis []*Combination) {
	combis = make([]*Combination, 0, 49)
	combis = append(combis, createCombiWedge(cardsNo)...)
	combis = append(combis, createCombiPhalanx(cardsNo)...)
	combis = append(combis, createCombiBattalion()...)
	combis = append(combis, createCombiSkirmish(cardsNo)...)
	combis = append(combis, createCombiHost())
	for i, c := range combis {
		c.Rank = i + 1
	}
	return combis
}
func createCombiHost() *Combination {
	c := &Combination{
		Formation:  card.FHost,
		TieBreaker: tbStrenght,
	}
	return c
}
func createCombiWedge(cardsNo int) []*Combination {
	combis := make([]*Combination, 0, 10+1-cardsNo)
	for strenght := 10; strenght >= cardsNo; strenght-- {
		combi := Combination{
			Formation:  card.FWedge,
			Troops:     make(map[int][]card.Troop),
			TieBreaker: tbRank,
		}

		for color := 1; color < 7; color++ {
			cardixs := make([]card.Troop, 0, cardsNo)
			for i := strenght; i > strenght-cardsNo; i-- {
				cardixs = append(cardixs, card.Troop((color-1)*10+i))
				if color == 1 {
					combi.Strength = combi.Strength + i
				}
			}
			combi.Troops[color] = cardixs
		}
		combis = append(combis, &combi)
	}
	return combis

}
func createCombiPhalanx(cardsNo int) []*Combination {
	combis := make([]*Combination, 0, 10)
	//Phalanx
	for str := 10; str > 0; str-- {
		combi := Combination{
			Formation:  card.FPhalanx,
			Strength:   str * cardsNo,
			Troops:     make(map[int][]card.Troop),
			TieBreaker: tbRank,
		}
		troops := make([]card.Troop, 0, 6)
		for color := 1; color < 7; color++ {
			troops = append(troops, card.Troop((color-1)*10+str))
		}
		combi.Troops[card.COLNone] = troops
		combis = append(combis, &combi)
	}
	return combis
}
func createCombiBattalion() (combis []*Combination) {
	allStrenghts := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	troops := make(map[int][]card.Troop)
	for color := 1; color <= card.NOColors; color++ {
		cardixs := make([]card.Troop, 0, 10)
		for _, str := range allStrenghts {
			cardixs = append(cardixs, card.Troop((color-1)*10+str))
		}
		troops[color] = cardixs
	}
	combis = make([]*Combination, 2)
	for i := range combis {
		combi := Combination{
			Formation:  card.FBattalion,
			Strength:   0,
			Troops:     troops,
			TieBreaker: tbStrenght,
		}
		combis[i] = &combi
	}
	combis[1].TieBreaker = tbNone
	return combis
}
func createCombiSkirmish(cardsNo int) (combis []*Combination) {
	combis = make([]*Combination, 0, 10+1-cardsNo)

	for str := 10; str >= cardsNo; str-- {
		combi := Combination{
			Formation:  card.FSkirmish,
			Troops:     make(map[int][]card.Troop),
			TieBreaker: tbRank,
		}
		baseixs := make([]int, 0, cardsNo)
		strenght := 0
		for i := str; i > str-cardsNo; i-- {
			baseixs = append(baseixs, i)
			strenght = strenght + i
		}
		combi.Strength = strenght
		cardixs := make([]card.Troop, 0, cardsNo*6)
		for color := 1; color < 7; color++ {
			for _, baseix := range baseixs {
				cardixs = append(cardixs, card.Troop((color-1)*10+baseix))
			}
		}
		combi.Troops[card.COLNone] = cardixs

		combis = append(combis, &combi)
	}
	return combis
}

//LastFormationRank returns last rank for a formation.
func LastFormationRank(formation card.Formation, formationSize int) (rank int) {
	combinations := combinations3
	if formationSize == 4 {
		combinations = combinations4
	}
	for _, c := range combinations {
		if c.Formation.Value < formation.Value {
			rank = c.Rank - 1
			break
		}
	}
	if rank == 0 {
		rank = len(combinations) + 1
	}
	return rank
}

//TieBreaker the tiebreaker rule of a combination.
type TieBreaker uint8

//IsRank returns true if rank is the tiebreaker.
//first to this rank wins.
func (t TieBreaker) IsRank() bool {
	return t == tbRank
}

//IsStrenght returns true if strenght is the tiebreaker.
//first to strenght wins.
func (t TieBreaker) IsStrenght() bool {
	return t == tbStrenght
}

//IsNone returns true if no tiebreaker rule
//this rank is not supposed to be combared
//with a equal rank.
func (t TieBreaker) IsNone() bool {
	return t == tbNone
}
func (t TieBreaker) String() string {
	return tbNames[int(t)]
}
