package machine

import (
	"testing"

	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
)

func TestMpos(t *testing.T) {
	expMpos := testCreateMpos()
	gamePos := bat.NewGamePos()
	gamePos.Hands[0].Draw(1)
	testSet(expMpos, 1, CardPosAll.HandLegal)
	gamePos.Hands[0].Draw(2)
	testSet(expMpos, 2, CardPosAll.HandLegal)
	gamePos.Hands[0].Draw(3)
	testSet(expMpos, 3, CardPosAll.HandLegal)
	gamePos.Hands[0].Draw(4)
	testSet(expMpos, 4, CardPosAll.HandLegal)
	gamePos.Hands[0].Draw(5)
	testSet(expMpos, 5, CardPosAll.HandLegal)
	gamePos.Hands[0].Draw(6)
	testSet(expMpos, 6, CardPosAll.HandLegal)
	gamePos.Hands[0].Draw(cards.TCRedeploy)
	testSet(expMpos, cards.TCRedeploy, CardPosAll.Hand)

	gamePos.Hands[1].Draw(7)
	gamePos.Hands[1].Draw(8)
	gamePos.Hands[1].Draw(9)
	gamePos.Hands[1].Draw(10)
	gamePos.Hands[1].Draw(11)
	gamePos.Hands[1].Draw(12)
	gamePos.Hands[1].Draw(cards.TCAlexander)
	testSet(expMpos, pOppHandTroopsNo, 6)
	testSet(expMpos, pOppHandTacsNo, 1)

	gamePos.Dishs[0].DishCard(13)
	testSet(expMpos, 13, CardPosAll.DishBot)
	gamePos.Dishs[0].DishCard(cards.TCTraitor)
	testSet(expMpos, cards.TCTraitor, CardPosAll.DishBot)
	gamePos.Dishs[1].DishCard(14)
	testSet(expMpos, 14, CardPosAll.DishOpp)
	gamePos.Dishs[1].DishCard(cards.TCDeserter)
	testSet(expMpos, cards.TCDeserter, CardPosAll.DishOpp)

	gamePos.Flags[0].Set(cards.TCFog, 0)
	testSet(expMpos, cards.TCFog, CardPosAll.FlagBot[0])
	gamePos.Flags[0].Set(20, 0)
	testSet(expMpos, 20, CardPosAll.FlagBot[0])
	gamePos.Flags[0].Set(30, 0)
	testSet(expMpos, 30, CardPosAll.FlagBot[0])
	gamePos.Flags[0].Set(40, 0)
	testSet(expMpos, 40, CardPosAll.FlagBot[0])
	gamePos.Flags[0].Set(21, 1)
	testSet(expMpos, 21, CardPosAll.FlagOpp[0])
	gamePos.Flags[0].Players[0].Won = true

	testSet(expMpos, pFlag, ConePos.Bot)

	gamePos.Flags[8].Set(cards.TCMud, 1)
	testSet(expMpos, cards.TCMud, CardPosAll.FlagOpp[8])
	gamePos.Flags[8].Set(29, 1)
	testSet(expMpos, 29, CardPosAll.FlagOpp[8])
	gamePos.Flags[8].Set(39, 1)
	testSet(expMpos, 39, CardPosAll.FlagOpp[8])
	gamePos.Flags[8].Set(49, 1)
	testSet(expMpos, 49, CardPosAll.FlagOpp[8])
	gamePos.Flags[8].Set(59, 1)
	testSet(expMpos, 59, CardPosAll.FlagOpp[8])
	gamePos.Flags[8].Set(22, 0)
	testSet(expMpos, 22, CardPosAll.FlagBot[8])
	gamePos.Flags[8].Players[1].Won = true
	testSet(expMpos, pFlag+8, ConePos.Opp)

	gamePos.Flags[4].Set(25, 0)
	testSet(expMpos, 25, CardPosAll.FlagBot[4])
	gamePos.Flags[4].Set(26, 1)
	testSet(expMpos, 26, CardPosAll.FlagOpp[4])

	moves := make(map[int][]bat.Move)
	move := make([]bat.Move, 1)

	move[0] = *bat.NewMoveCardFlag(2)

	moves[1] = move
	moves[2] = move
	moves[3] = move
	moves[4] = move
	moves[5] = move
	moves[6] = move

	gamePos.Turn.MovesHand = moves
	moveix := 22
	testSet(expMpos, 0, uint8(moveix))
	cardix := 1
	testSetMove(expMpos, pMoveFirstCard, uint8(cardix))
	testSetMove(expMpos, pMoveFirstCardDest, MoveDestAll.Flag[2])
	mpos := CreatePos(gamePos, move[0], nil, false, cardix, 0, 0, moveix)
	testCompare(mpos, expMpos, t)

	tacs := []int{cards.TCDarius}
	troops := []int{23}
	scoutR := *bat.NewMoveScoutReturn(tacs, troops)
	testSet(expMpos, pScoutRBotFirstCard, cards.TCDarius)
	testSet(expMpos, pScoutRBotSecondCard, 23)
	mpos = CreatePos(gamePos, move[0], scoutR, false, cardix, 0, 0, moveix)
	testCompare(mpos, expMpos, t)

	testSet(expMpos, pScoutRBotFirstCard, 0)
	testSet(expMpos, pScoutRBotSecondCard, 0)
	testSet(expMpos, pOppKnowDeckTacsNo, 1)
	testSet(expMpos, pOppKnowDeckTroopsNo, 1)
	mpos = CreatePos(gamePos, move[0], scoutR, false, cardix, 0, 1, moveix)
	testCompare(mpos, expMpos, t)

	testSet(expMpos, pOppKnowDeckTacsNo, 1)
	testSet(expMpos, pOppKnowDeckTroopsNo, 0)
	testSet(expMpos, pOppKnowHand, 1)

	tacs = []int{cards.TCDarius}
	troops = []int{1}
	scoutR = *bat.NewMoveScoutReturn(tacs, troops)

}
func testCompare(mpos, expMpos MPos, t *testing.T) {
	if len(mpos) != len(expMpos) {
		t.Errorf("Length mismatch exp %v got %v", len(expMpos), len(mpos))
	} else {
		for i, v := range mpos {
			if v != expMpos[i] {
				t.Errorf("Mismatch on element %v exp %v got %v", i, expMpos[i], v)
			}
		}
	}
}
func testCreateMpos() MPos {
	mpos := make([]uint8, mPosNoBytes)
	for i := pCard; i <= cards.NOTac+cards.NOTroop; i++ {
		mpos[i] = CardPosAll.Deck
	}
	return mpos
}
func testSet(mpos []uint8, i int, v uint8) {
	mpos[i] = v
}
func testSetMove(mpos []uint8, i int, v uint8) {
	mpos[len(mpos)-4+i] = v
}
