package combi

import (
	"fmt"
	"github.com/rezder/go-battleline/battleline/cards"
)

var Combinations3, Combinations4 []*Combination

func init() {
	Combinations3 = createCombi(3)
	Combinations4 = createCombi(4)
}
func Combinations(size int) []*Combination {
	if size == 4 {
		return Combinations4
	}
	return Combinations3
}
func CombinationsMud(isMud bool) []*Combination {
	if isMud {
		return Combinations4
	}
	return Combinations3
}

//Combination a battleline formation and strength
type Combination struct {
	Rank      int
	Formation cards.Formation
	Strength  int
	//Troops is all the cards that can be used to create the formation.
	Troops map[int][]int
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
	for i, c := range combis {
		c.Rank = i + 1
	}
	return combis
}
func createCombiWedge(cardsNo int) []*Combination {
	combis := make([]*Combination, 0, 10+1-cardsNo)
	for value := 10; value >= cardsNo; value-- {
		combi := Combination{
			Formation: cards.FWedge,
			Troops:    make(map[int][]int),
		}

		for color := 1; color < 7; color++ {
			cardixs := make([]int, 0, cardsNo)
			for i := value; i > value-cardsNo; i-- {
				cardixs = append(cardixs, (color-1)*10+i)
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
	for value := 10; value > 0; value-- {
		combi := Combination{
			Formation: cards.FPhalanx,
			Strength:  value * cardsNo,
			Troops:    make(map[int][]int),
		}
		cardixs := make([]int, 0, 6)
		for color := 1; color < 7; color++ {
			cardixs = append(cardixs, (color-1)*10+value)
		}
		combi.Troops[cards.COLNone] = cardixs
		combis = append(combis, &combi)
	}
	return combis
}
func createCombiBattalion(cardsNo int) (combis []*Combination) {

	//Battalion Order
	allCards := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
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
			Formation: cards.FBattalion,
			Strength:  sum,
			Troops:    make(map[int][]int),
		}
		/*fac := math.FactorSum(allCards, sum, cardsNo, true)
		valueSet := make(map[int]bool)
		for _, ixs := range fac {
			for _, ix := range ixs {
				valueSet[allCards[ix]] = true
			}
		}
		*/
		for color := 1; color < 7; color++ {
			cardixs := make([]int, 0, 10)
			for _, value := range allCards {
				cardixs = append(cardixs, (color-1)*10+value)
			}
			combi.Troops[color] = cardixs
		}
		combis = append(combis, &combi)
	}
	return combis
}
func createCombiSkirmish(cardsNo int) (combis []*Combination) {
	combis = make([]*Combination, 0, 10+1-cardsNo)

	for value := 10; value >= cardsNo; value-- {
		combi := Combination{
			Formation: cards.FSkirmish,
			Troops:    make(map[int][]int),
		}
		baseixs := make([]int, 0, cardsNo)
		strenght := 0
		for i := value; i > value-cardsNo; i-- {
			baseixs = append(baseixs, i)
			strenght = strenght + i
		}
		combi.Strength = strenght
		cardixs := make([]int, 0, cardsNo*6)
		for color := 1; color < 7; color++ {
			for _, baseix := range baseixs {
				cardixs = append(cardixs, (color-1)*10+baseix)
			}
		}
		combi.Troops[cards.COLNone] = cardixs

		combis = append(combis, &combi)
	}
	return combis
}
func LastFormationRank(formation cards.Formation, formationSize int) (rank int) {
	combinations := Combinations3
	if formationSize == 4 {
		combinations = Combinations4
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
