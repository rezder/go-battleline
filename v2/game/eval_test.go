package game

import (
	"encoding/gob"
	"encoding/json"
	"github.com/rezder/go-battleline/v2/game/card"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//testEval tests the eval function.
//WARNING Expects troops to be sorted
type testEval struct {
	Troops    []card.Troop
	Morales   []card.Morale
	IsMud     bool
	IsFog     bool
	FormValue int
	Strenght  int
	Info      string
}

const testDir = "test"

type enCoder interface {
	Encode(v interface{}) error
}
type deCoder interface {
	Decode(v interface{}) error
}

func TestEval(t *testing.T) {
	fileName := "evals.json"
	var savedEvals []*testEval
	fp := filepath.Join(testDir, fileName)
	err := testLoadFile(fp, &savedEvals)
	if err != nil {
		t.Errorf("error loading: %v", err)
	}
	for i, tv := range savedEvals {
		form, strenght := eval(tv.Troops, tv.Morales, tv.IsMud, tv.IsFog)
		if form.Value != tv.FormValue || strenght != tv.Strenght {
			t.Errorf("Test Info: %v, element %v failed. Expect formation value: %v got %v. Expected strenght: %v got %v",
				tv.Info, i, tv.FormValue, form.Value, tv.Strenght, strenght)
		}
	}
}
func TestFlag(t *testing.T) {
	fileName := "flags.json"
	var tests []*testFlag
	fp := filepath.Join(testDir, fileName)
	err := testLoadFile(fp, &tests)
	if err != nil {
		t.Errorf("error loading: %v", err)
	}
	for testix, tf := range tests {
		flag := tf.createFlag()
		deck := tf.createDeck(flag)
		testFlagEstimate(0, flag, deck, testix, t)
		testFlagsPrint(0, flag, deck, tf, testix, t)
		testFlagsPrint(1, flag, deck, tf, testix, t)
	}
}
func testFlagEstimate(player int, flag *Flag, deckTroops []card.Troop, testix int, t *testing.T) {
	troops := flag.Players[player].Troops
	morales := flag.Players[player].Morales
	if len(troops) > 1 && !flag.HasFormation(player) {

		eForm, eStrenght := estimateFormation(
			flag.IsFog,
			flag.IsMud,
			troops,
			morales)
		isClaim, _ := isFlagClaimableSim(
			eForm,
			eStrenght,
			flag.IsFog,
			flag.IsMud,
			troops,
			morales,
			deckTroops,
		)
		if isClaim {
			t.Errorf("Test element %v failed. Estimate formation: %v,strenght: %v,must never be topped.", testix, eForm, eStrenght)
		}
	}
}

func testFlagsPrint(player int, flag *Flag, deckTroops []card.Troop, tf *testFlag, testix int, t *testing.T) {
	isClaim := false
	if flag.HasFormation(player) {
		isClaim, _ = flag.IsClaimable(player, deckTroops)
	}
	if isClaim != tf.IsClaimables[player] {
		t.Logf("flag: %v\n,deck: %v", flag, deckTroops)
		t.Errorf("Test element: %v testing :%v failed. Expect claim: %v got %v.", testix, tf.Info, tf.IsClaimables[player], isClaim)
	}
}

type testFlag struct {
	FlagCardixs  [2][]int
	IsMud        bool
	IsFog        bool
	IsNegDef     bool
	DeckTroops   []card.Troop
	IsClaimables [2]bool
	Info         string
}

func (t *testFlag) createFlag() (flag *Flag) {
	flag = new(Flag)
	flag.IsFog = t.IsFog
	flag.IsMud = t.IsMud
	for player := 0; player < 2; player++ {
		for _, cardix := range t.FlagCardixs[player] {
			cardMove := card.Move(cardix)
			switch {
			case cardMove.IsTroop():
				flag.Players[player].Troops = appendSortedTroop(flag.Players[player].Troops, card.Troop(cardMove))
			case cardMove.IsMorale():
				flag.Players[player].Morales = append(flag.Players[player].Morales, card.Morale(cardMove))
			}
		}
	}
	return flag
}
func (t *testFlag) createDeck(flag *Flag) (deckTroops []card.Troop) {
	if t.IsNegDef {
		noRemoves := len(t.DeckTroops) + len(flag.Players[0].Troops) + len(flag.Players[1].Troops)
		removeTroops := make([]card.Troop, noRemoves)
		copy(removeTroops, t.DeckTroops)
		copy(removeTroops[len(t.DeckTroops):], flag.Players[0].Troops)
		copy(removeTroops[noRemoves-len(flag.Players[1].Troops):], flag.Players[1].Troops)
		deckSize := card.NOTroop - len(removeTroops)
		deckTroops = make([]card.Troop, 0, deckSize)
		for cardix := 1; cardix <= card.NOTroop; cardix++ {
			troop := card.Troop(cardix)
			isAdd := true
			for _, removeTroop := range removeTroops {
				if removeTroop == troop {
					isAdd = false
					break
				}
			}
			if isAdd {
				deckTroops = append(deckTroops, troop)
			}
		}
	} else {
		deckTroops = t.DeckTroops
	}
	return deckTroops
}

func testsSaveFile(filePath string, v interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var coder enCoder
	if strings.HasSuffix(filePath, "json") {
		coder = json.NewEncoder(file)
	} else {
		coder = gob.NewEncoder(file)
	}
	err = coder.Encode(v)
	return err
}
func testLoadFile(filePath string, ts interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var coder deCoder
	if strings.HasSuffix(filePath, "json") {
		coder = json.NewDecoder(file)
	} else {
		coder = gob.NewDecoder(file)
	}
	err = coder.Decode(ts)
	return err
}
func TestConeMoves(t *testing.T) {
	moves := coneCombiMoves([]int{0, 1, 2, 3, 4, 5, 6, 7, 8}, 0)
	for moveix, move := range moves {
		noBpMove := len(move.Moves)
		for i, bpmove := range move.Moves {
			if i < noBpMove-1 {
				if bpmove.Index > move.Moves[i+1].Index {
					t.Errorf("Flag index must be in increasing order moveix: %v", moveix)
					//This is to get consisten moves and easy to replica a move
				}
			}
		}
	}
}
