package combi

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/game/card"
)

const (
	//COLNone none color
	COLNone = 0
)

var combinations3, combinations4 []*Combination

func init() {
	combinations3 = createCombi(3)
	combinations4 = createCombi(4)
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

// HostRank returns host rank
func HostRank(size int) int {
	return len(Combinations(size))
}

//Combination a battleline formation and strength
type Combination struct {
	Rank      int
	Formation card.Formation
	Strength  int
	//Troops is all the cards that can be used to create the formation.
	//per color
	Troops map[int][]card.Troop
}

func (c *Combination) String() string {
	if c == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{Rank:%v Strength:%v Formation:%v Troops:%v",
		c.Rank, c.Strength, c.Formation, c.Troops)
}

//createCombi create all the possible combinations for the specified number of
//cards 3 or 4.
func createCombi(cardsNo int) (combis []*Combination) {
	combis = make([]*Combination, 0, 49)
	combis = append(combis, createCombiWedge(cardsNo)...)
	combis = append(combis, createCombiPhalanx(cardsNo)...)
	combis = append(combis, createCombiBattalion(cardsNo)...)
	combis = append(combis, createCombiSkirmish(cardsNo)...)
	combis = append(combis, createCombiHost())
	for i, c := range combis {
		c.Rank = i + 1
	}
	return combis
}
func createCombiHost() *Combination {
	c := &Combination{
		Formation: card.FHost,
	}
	return c
}
func createCombiWedge(cardsNo int) []*Combination {
	combis := make([]*Combination, 0, 10+1-cardsNo)
	for strenght := 10; strenght >= cardsNo; strenght-- {
		combi := Combination{
			Formation: card.FWedge,
			Troops:    make(map[int][]card.Troop),
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
			Formation: card.FPhalanx,
			Strength:  str * cardsNo,
			Troops:    make(map[int][]card.Troop),
		}
		troops := make([]card.Troop, 0, 6)
		for color := 1; color < 7; color++ {
			troops = append(troops, card.Troop((color-1)*10+str))
		}
		combi.Troops[COLNone] = troops
		combis = append(combis, &combi)
	}
	return combis
}
func createCombiBattalion(cardsNo int) (combis []*Combination) {

	//Battalion Order
	allStrenghts := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var maxsum int
	if cardsNo == 3 {
		maxsum = 27 //J10+10+7 higest no straight flush
	} else {
		maxsum = 36 //J10+10+9+6 J10,10,j8,8
	}
	minsum := 0
	for i := 1; i < cardsNo+1; i++ {
		minsum = minsum + i
	}
	combis = make([]*Combination, 0, maxsum-minsum)
	for sum := maxsum; sum > minsum; sum-- {
		combi := Combination{
			Formation: card.FBattalion,
			Strength:  sum,
			Troops:    make(map[int][]card.Troop),
		}
		//TODO I do not know why I did this?
		/*fac := math.FactorSum(allCards, sum, cardsNo, true)
		strSet := make(map[int]bool)
		for _, ixs := range fac {
			for _, ix := range ixs {
				strSet[allStrenghts[ix]] = true
			}
		}
		*/
		for color := 1; color < 7; color++ {
			cardixs := make([]card.Troop, 0, 10)
			for _, str := range allStrenghts {
				cardixs = append(cardixs, card.Troop((color-1)*10+str))
			}
			combi.Troops[color] = cardixs
		}
		combis = append(combis, &combi)
	}
	return combis
}
func createCombiSkirmish(cardsNo int) (combis []*Combination) {
	combis = make([]*Combination, 0, 10+1-cardsNo)

	for str := 10; str >= cardsNo; str-- {
		combi := Combination{
			Formation: card.FSkirmish,
			Troops:    make(map[int][]card.Troop),
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
		combi.Troops[COLNone] = cardixs

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
