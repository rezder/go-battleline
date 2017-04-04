package machine

import (
	"fmt"

	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
	gflag "github.com/rezder/go-battleline/battleline/flag"
)

type MPos []uint8

const (
	mPosNoBytes = 93
)

func NewMPos() MPos {
	mpos := make([]uint8, mPosNoBytes)
	return mpos
}
func CreatePos(pos *bat.GamePos, move, scoutRMove bat.Move, passMove bool, moveCardix, mover, scoutMover, gameMoveix int) MPos {
	mpos := setCardsPos(pos, mover)
	mpos[pMoveix] = uint8(gameMoveix)
	setConesPos(pos.Flags, mover, mpos)
	setOppontFields(pos.Hands[pos.Opp()], mpos)
	if scoutMover == mover {
		setScoutReturnBotFields(scoutRMove, mpos)
	} else {
		setScoutReturnOppFields(scoutRMove, mpos)
	}
	if pos.MovePass {
		mpos[pPass] = 1
	}
	moveBytes := CreateMMove(moveCardix, move, passMove)
	for i := 1; i <= mMoveNoBytes; i++ {
		mpos[mPosNoBytes-i] = moveBytes[mMoveNoBytes-i]
	}
	return mpos
}
func setConesPos(flags [9]*gflag.Flag, mover int, mpos MPos) {
	intix := pFlag
	for flagix, flag := range flags {
		if !flag.Claimed() {
			mpos[intix+flagix] = ConePos.None
		} else {
			if flag.Players[mover].Won {
				mpos[intix+flagix] = ConePos.Bot
			} else {
				mpos[intix+flagix] = ConePos.Opp
			}
		}
	}
}
func setOppontFields(hand *bat.Hand, mpos MPos) {
	mpos[pOppHandTacsNo] = uint8(len(hand.Tacs))
	mpos[pOppHandTroopsNo] = uint8(len(hand.Troops))
}

func setScoutReturnOppFields(move bat.Move, mpos MPos) {

	scoutReturnMove, ok := move.(bat.MoveScoutReturn)
	if ok {
		tacHandNo, tacDeckNo := countCardsHandDeck(scoutReturnMove.Tac, mpos)
		troopHandNo, troopDeckNo := countCardsHandDeck(scoutReturnMove.Troop, mpos)
		mpos[pOppKnowDeckTacsNo] = uint8(tacDeckNo)
		mpos[pOppKnowDeckTroopsNo] = uint8(troopDeckNo)
		if tacHandNo+troopHandNo > 0 {
			cardixs := make([]int, 0, 2)
			cardixs = append(cardixs, scoutReturnMove.Tac...)
			cardixs = append(cardixs, scoutReturnMove.Troop...)
			for i, cardix := range cardixs {
				mpos[pOppKnowHand+i] = uint8(cardix)
			}
		}

	}
}
func countCardsHandDeck(cardixs []int, mpos MPos) (handNo, deckNo int) {
	for _, cardix := range cardixs {
		if mpos[cardix] == CardPosAll.Deck {
			deckNo = deckNo + 1
		} else if mpos[cardix] == CardPosAll.Hand || mpos[cardix] == CardPosAll.HandLegal {
			handNo = handNo + 1
		}
	}
	return handNo, deckNo
}
func setScoutReturnBotFields(move bat.Move, mpos MPos) {

	scoutReturnMove, ok := move.(bat.MoveScoutReturn)
	if ok {
		if len(scoutReturnMove.Tac) == 1 {
			mpos[pScoutRBotFirstCard] = uint8(scoutReturnMove.Tac[0])
			if len(scoutReturnMove.Troop) == 1 {
				mpos[pScoutRBotSecondCard] = uint8(scoutReturnMove.Troop[0])
			}
		} else {
			if len(scoutReturnMove.Tac) == 2 {
				mpos[pScoutRBotFirstCard] = uint8(scoutReturnMove.Tac[0])
				mpos[pScoutRBotSecondCard] = uint8(scoutReturnMove.Tac[1])
			} else if len(scoutReturnMove.Troop) == 2 {
				mpos[pScoutRBotFirstCard] = uint8(scoutReturnMove.Troop[0])
				mpos[pScoutRBotSecondCard] = uint8(scoutReturnMove.Troop[1])
			}
		}
		scoutReturnBotClear(mpos, pScoutRBotFirstCard)
		scoutReturnBotClear(mpos, pScoutRBotSecondCard)
	}
}
func scoutReturnBotClear(mpos MPos, ix int) {
	if mpos[ix] != 0 {
		cardPos := mpos[mpos[ix]]
		if !(cardPos == CardPosAll.Deck || cardPos == CardPosAll.Hand || cardPos == CardPosAll.HandLegal) {
			mpos[ix] = 0
		}
	}
}
func setCardsPos(pos *bat.GamePos, mover int) MPos {
	m := NewMPos()
	startix := pCard - 1 //-1 is because cardix is not zero indexed
	for i := pCard; i <= cards.NOTac+cards.NOTroop; i++ {
		m[i] = CardPosAll.Deck
	}
	for _, cardix := range pos.Hands[mover].Troops {
		if cardix != 0 {
			m[startix+cardix] = CardPosAll.Hand
		}
	}
	for _, cardix := range pos.Hands[mover].Tacs {
		if cardix != 0 {
			m[startix+cardix] = CardPosAll.Hand
		}
	}
	for cardix := range pos.MovesHand {
		m[startix+cardix] = CardPosAll.HandLegal
	}

	for flagix, flag := range pos.Flags {
		for playerix, player := range flag.Players {
			isOpponent := mover != playerix
			setCardPosFlag(m, player.Env[:], flagix, isOpponent, startix)
			setCardPosFlag(m, player.Troops[:], flagix, isOpponent, startix)
		}
	}
	for playerix, dish := range pos.Dishs {
		isOpponent := mover != playerix
		setCardPosDish(m, dish.Tacs, isOpponent, startix)
		setCardPosDish(m, dish.Troops, isOpponent, startix)
	}
	return m
}
func setCardPosDish(mpos MPos, cardixs []int, isOpponent bool, startix int) {
	for _, cardix := range cardixs {
		if cardix != 0 {
			if isOpponent {
				mpos[startix+cardix] = CardPosAll.DishOpp
			} else {
				mpos[startix+cardix] = CardPosAll.DishBot
			}
		}
	}
}
func setCardPosFlag(mpos MPos, cardixs []int, flagix int, isOpponent bool, startix int) {
	for _, cardix := range cardixs {
		if cardix != 0 {
			if isOpponent {
				mpos[startix+cardix] = CardPosAll.FlagOpp[flagix]
			} else {
				mpos[startix+cardix] = CardPosAll.FlagBot[flagix]
			}
		}
	}
}

type Domain interface {
	Name() string
	TxtToValue(string) uint8
	ValueToTxt(uint8) string
	DomValues() []uint8
}
type SimpleDom struct {
	name      string
	values    []uint8
	txtValues []string
}

func (dom *SimpleDom) DomValues() []uint8 {
	return dom.values
}
func (dom *SimpleDom) Name() string {
	return dom.name
}
func (dom *SimpleDom) ValueToTxt(value uint8) string {
	return domValueToTxt(dom.values, dom.txtValues, value)
}
func domValueToTxt(values []uint8, txts []string, value uint8) string {
	ix, ok := domValueIndex(value, values)
	if !ok {
		panic(fmt.Sprintf("Domain value %v does not exist", value))
	}
	return txts[ix]
}
func domValueIndex(value uint8, values []uint8) (ix int, ok bool) {
	for i, v := range values {
		if v == value {
			ok = true
			ix = i
			break
		}
	}
	return ix, ok
}
func (dom *SimpleDom) TxtToValue(txt string) uint8 {
	return domTxtToValue(dom.values, dom.txtValues, txt)
}
func domTxtIndex(txtValues []string, txtValue string) (ix int, ok bool) {
	for i, tv := range txtValues {
		if tv == txtValue {
			ok = true
			ix = i
			break
		}
	}
	return ix, ok
}
func domTxtToValue(values []uint8, txtValues []string, txtValue string) (value uint8) {
	ix, ok := domTxtIndex(txtValues, txtValue)
	if !ok {
		panic(fmt.Sprintf("Domain text value %v does not exit", txtValue))
	}
	return values[ix]
}

var (
	CardPosAll *CardPosAllSingleton
	ConePos    *ConePosSingleton
)

func init() {
	CardPosAll = newCardPosAllSingleton()
	ConePos = newConePosSingleton()
}

type CardPosAllSingleton struct {
	SimpleDom
	DishBot   uint8
	FlagBot   [9]uint8
	DishOpp   uint8
	FlagOpp   [9]uint8
	HandLegal uint8
	Hand      uint8
	Deck      uint8
}

func newCardPosAllSingleton() (cp *CardPosAllSingleton) {
	cp = new(CardPosAllSingleton)
	cp.name = "Card Positions"
	cp.values = []uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22}
	cp.txtValues = make([]string, 23)
	cp.DishBot = 0
	cp.txtValues[0] = "DishBot"
	cp.DishOpp = 10
	cp.txtValues[10] = "DishOpp"
	for i := 0; i < 9; i++ {
		botix := i + 1
		oppix := i + 11
		cp.FlagBot[i] = uint8(botix)
		cp.FlagOpp[i] = uint8(oppix)
		cp.txtValues[botix] = fmt.Sprint("Flag%vBot", i+1)
		cp.txtValues[oppix] = fmt.Sprint("Flag%vOpp", i+1)
	}
	cp.HandLegal = 20
	cp.txtValues[20] = "HandLegal"
	cp.Hand = 21
	cp.txtValues[21] = "Hand"
	cp.Deck = 22
	cp.txtValues[22] = "Deck"

	return cp
}
func (cp *CardPosAllSingleton) Flagix(pos uint8) (ix int, ok bool) {
	ok = false
	if pos > 0 && pos < 11 {
		ix = int(pos) - 1
		ok = true
	} else if pos > 10 && pos < 21 {
		ix = int(pos) - 11
		ok = true
	}
	return ix, ok
}

type CardPosDom struct {
	name   string
	values []uint8
}

func NewCardPosDom(values []uint8) (cp *CardPosDom) {
	cp = new(CardPosDom)
	cp.values = values
	return cp
}
func (cp *CardPosDom) Name() string {
	return cp.name
}
func (cp *CardPosDom) TxtToValue(txtValue string) uint8 {
	value := CardPosAll.TxtToValue(txtValue)
	_, ok := domValueIndex(value, cp.values)

	if !ok {
		panic(fmt.Sprintf("Domain text value %v does not exist", txtValue))
	}
	return value
}
func (cp *CardPosDom) ValueToTxt(value uint8) string {
	_, ok := domValueIndex(value, cp.values)
	if !ok {
		panic(fmt.Sprintf("Domain value %v does not exist", value))
	}
	return CardPosAll.ValueToTxt(value)
}
func (cp *CardPosDom) DomValues() []uint8 {
	return cp.values
}

type ConePosSingleton struct {
	SimpleDom
	None uint8
	Bot  uint8
	Opp  uint8
}

func newConePosSingleton() (cp *ConePosSingleton) {
	cp = new(ConePosSingleton)
	cp.txtValues = make([]string, 3)
	cp.name = "Cone Positions"
	cp.None = 0
	cp.txtValues[0] = "None"
	cp.Bot = 1
	cp.txtValues[1] = "Bot"
	cp.Opp = 2
	cp.txtValues[2] = "Opp"
	cp.values = []uint8{0, 1, 2}
	return cp
}

type MPosFld struct {
	Fld
	MposIx int
}

func MPosCreateFeatureFlds() (flds []Fld) {
	flds = make([]Fld, 92)
	ix := 0
	guileDom := NewCardPosDom([]uint8{CardPosAll.Deck, CardPosAll.DishBot, CardPosAll.DishOpp, CardPosAll.Hand, CardPosAll.HandLegal})
	var dom Domain
	for i := 0; i < cards.NOTac+cards.NOTroop; i++ {
		dom = CardPosAll
		cardix := i + 1
		if cards.IsGuile(cardix) {
			dom = guileDom
		}
		fld := MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: fmt.Sprintf("Card %v", cardix),
			Domain: dom}}
		flds[ix] = fld
		ix = ix + 1
	}
	for i := 0; i < bat.NOFlags; i++ {
		fld := MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: fmt.Sprintf("Flag %v", i+1),
			Domain: ConePos}}
		flds[ix] = fld
		ix = ix + 1
	}
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &ValueFld{Name: "Opp. Hand No. Troops", Scale: 7}}
	ix = ix + 1
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &ValueFld{Name: "Opp. Hand No. Tactics", Scale: 4}}
	ix = ix + 1
	noScoutDomain := newCardDomain("Scout Return Cards", []uint8{uint8(cards.TCScout)})
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Scout Return Bot First Card",
		Domain: noScoutDomain}}
	ix = ix + 1
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Scout Return Bot Second Card",
		Domain: noScoutDomain}}
	ix = ix + 1
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &ValueFld{Name: "Opp. Know Deck Tactics", Scale: 2}}
	ix = ix + 1
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &ValueFld{Name: "Opp. Know Deck Troops", Scale: 2}}
	ix = ix + 1
	for i := 0; i < 2; i++ {
		fld := MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: fmt.Sprintf("Opp. Know Card On Hand %v", i+1),
			Domain: noScoutDomain}}
		flds[ix] = fld
		ix = ix + 1
	}
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &ValueFld{Name: "Pass Possible", Scale: 1}}

	allCardsDomain := newCardDomain("All Cards", nil)
	ix = ix + 1
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Move FirstCard", Domain: allCardsDomain}}
	ix = ix + 1
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Move FirstCard Destination", Domain: MoveDestAll}}
	ix = ix + 1
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Move SecondCard", Domain: noScoutDomain}}
	ix = ix + 1
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Move SecondCard Destination", Domain: MoveDestAll}}

	return flds
}

type CardDomain struct {
	removeixs []uint8
	name      string
	colors    [6]string
	values    []uint8
}

func newCardDomain(name string, removeixs []uint8) (cd *CardDomain) {
	cd = new(CardDomain)
	cd.removeixs = removeixs
	cd.name = name
	cd.colors = [6]string{"Green", "Red", "Purpel", "Yellow", "Blue", "Orange"}
	cd.values = make([]uint8, cards.NOTac+cards.NOTroop+1-len(cd.removeixs))
	removed := 0
	for i := 0; i < cards.NOTac+cards.NOTroop+1; i++ {
		remove := false
		for _, v := range cd.removeixs {
			if v == uint8(i) {
				remove = true
				removed = removed + 1
			}
		}
		if !remove {
			cd.values[i-removed] = uint8(i)
		}
	}
	return cd
}
func (cd *CardDomain) Name() (name string) {
	return cd.name
}

func (cd *CardDomain) TxtToValue(txt string) (value uint8) {
	found := false
	for _, v := range cd.values {
		if cd.ValueToTxt(v) == txt {
			found = true
			value = v
			break
		}
	}
	if !found {
		panic(fmt.Sprintf("Card text value %v does not exist.", txt))
	}
	return value
}
func (cd *CardDomain) ValueToTxt(v uint8) (txt string) {
	if v == 0 {
		txt = "None"
	} else {
		troop, ok := cards.DrTroop(int(v))
		if ok {
			txt = fmt.Sprintf("%v%v", cd.colors[troop.Color()-1], troop.Value())
		} else {
			tac, ok := cards.DrTactic(int(v))
			if !ok {
				panic(fmt.Sprintf("Card value %v does not exist.", v))
			}
			txt = tac.Name()
		}
	}
	return txt
}
func (cd *CardDomain) DomValues() (values []uint8) {
	return cd.values
}

var (
	pMoveix              = 0  //0 1
	pCard                = 1  //1 66*23+4*5=1538
	pFlag                = 71 //1539 27
	pOppHandTroopsNo     = 80 //1566 1
	pOppHandTacsNo       = 81 //1567 1
	pScoutRBotFirstCard  = 82 //1568 70
	pScoutRBotSecondCard = 83 //1638 70
	pOppKnowDeckTacsNo   = 84 //1708 1
	pOppKnowDeckTroopsNo = 85 //1709 1
	pOppKnowHand         = 86 //1710 2*70=140
	pPass                = 88 //1850 1
	//pMoveFirstCard       = 89 //1851 71 does not include special SPCClaimFlag or SPCDeck
	//pMoveFirstCardDest   = 90 //1922 11 pos:dish,flag,deck
	//pMoveSecondCard      = 91 //1933 70 no scout
	//pMoveSecondCardDest  = 92 //2003 11   pos:dish,flag,deck. Features 2013
)
