package machine

import (
	"fmt"
	botflag "github.com/rezder/go-battleline/battbot/flag"
	bot "github.com/rezder/go-battleline/battbot/gamepos"
	bat "github.com/rezder/go-battleline/battleline"
	pub "github.com/rezder/go-battleline/battserver/publist"

	"github.com/rezder/go-battleline/battleline/cards"
	gflag "github.com/rezder/go-battleline/battleline/flag"
	"strconv"
)

// MPos a machine position.
type MPos []uint8

const (
	mPosNoBytes = 93
)

// NewMPos Create new machine position all all card position is zero
func NewMPos() MPos {
	mpos := make([]uint8, mPosNoBytes)
	return mpos
}

// CreatePosBot create a machine positions from a bot position.
// The positions are serialize.
func CreatePosBot(pos *bot.Pos) (data []byte, moves [][2]int) {
	mpos := setBotCardsPos(pos)
	mpos[pMoveix] = uint8(0)
	setBotConesPos(pos.Flags, mpos)
	setBotOppontFields(pos.Deck.OppTacNo(), pos.Deck.OppTroopNo(), mpos)
	setScoutReturnBotFields(pos.Deck.TfScoutReturnMoveTacs(), pos.Deck.TfScoutReturnMoveTroops(), mpos)
	noTacs, noTroops, handCardixs := pos.Deck.OppKnowns()
	setBotScoutReturnOppFields(noTacs, noTroops, handCardixs, mpos)
	if pos.Turn.MovesPass {
		mpos[pPass] = 1
	}
	var mmoves []MMove
	mmoves, moves = createMMovesBot(pos.Turn.MovesHand, pos.Turn.MovesPass)
	data = serializeBotPos(mpos, mmoves)
	return data, moves
}

// createMposBot create machine moves and coresponding game server moves.
func createMMovesBot(batMoves map[string][]bat.Move, pass bool) (mmoves []MMove, moves [][2]int) {
	moves = make([][2]int, 0, 40)
	mmoves = make([]MMove, 0, 40)
	for cardixTxt, cardMoves := range batMoves {
		cardix, _ := strconv.Atoi(cardixTxt)
		for i, move := range cardMoves {
			moves = append(moves, [2]int{cardix, i})
			mmoves = append(mmoves, CreateMMove(cardix, move, false))
		}
	}
	if pass {
		moves = append(moves, [2]int{0, pub.SMPass})
		mmoves = append(mmoves, NewMMove())
	}
	return mmoves, moves
}

// serializeBotPos join machine pos and moves and the serialize them.
func serializeBotPos(mPos MPos, mmoves []MMove) (data []byte) {
	data = make([]byte, 0, len(mmoves)*(mPosNoBytes-1))
	for _, mMove := range mmoves {
		newMove := AddMove(mPos, mMove)
		data = append(data, newMove[1:]...)
	}
	return data
}

// CreatePos create machine pos from game position
func CreatePos(pos *bat.GamePos, move, scoutRMove bat.Move, passMove bool, moveCardix, mover, scoutMover, gameMoveix int) MPos {
	mpos := setCardsPos(pos, mover)
	mpos[pMoveix] = uint8(gameMoveix)
	setConesPos(pos.Flags, mover, mpos)
	setOppontFields(pos.Hands[pos.Opp()], mpos)
	scoutReturnMove, ok := scoutRMove.(bat.MoveScoutReturn)
	if ok {
		if scoutMover == mover {
			setScoutReturnBotFields(scoutReturnMove.Tac, scoutReturnMove.Troop, mpos)
		} else {
			setScoutReturnOppFields(scoutReturnMove.Tac, scoutReturnMove.Troop, mpos)
		}
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
func setBotConesPos(flags [bat.NOFlags]*botflag.Flag, mpos MPos) {
	intix := pFlag
	for flagix, flag := range flags {
		if !flag.IsClaimed() {
			mpos[intix+flagix] = ConePos.None
		} else {
			if flag.Claimed == botflag.CLAIMPlay {
				mpos[intix+flagix] = ConePos.Bot
			} else {
				mpos[intix+flagix] = ConePos.Opp
			}
		}
	}
}
func setConesPos(flags [bat.NOFlags]*gflag.Flag, mover int, mpos MPos) {
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
func setBotOppontFields(tacNo, troopNo int, mpos MPos) {
	mpos[pOppHandTacsNo] = uint8(tacNo)
	mpos[pOppHandTroopsNo] = uint8(troopNo)
}
func setOppontFields(hand *bat.Hand, mpos MPos) {
	mpos[pOppHandTacsNo] = uint8(len(hand.Tacs))
	mpos[pOppHandTroopsNo] = uint8(len(hand.Troops))
}
func setBotScoutReturnOppFields(noTacs, noTroops int, cardixs []int, mpos MPos) {
	mpos[pOppKnowDeckTacsNo] = uint8(noTacs)
	mpos[pOppKnowDeckTroopsNo] = uint8(noTroops)
	i := 0
	for _, cardix := range cardixs {
		if mpos[cardix] == CardPosAll.Hand || mpos[cardix] == CardPosAll.HandLegal {
			mpos[pOppKnowHand+i] = uint8(cardix)
			i = i + 1
		}
	}
}
func setScoutReturnOppFields(tacs, troops []int, mpos MPos) {
	tacOnHandixs, tacDeckNo := countCardsHandDeck(tacs, mpos)
	troopOnHandixs, troopDeckNo := countCardsHandDeck(troops, mpos)
	mpos[pOppKnowDeckTacsNo] = uint8(tacDeckNo)
	mpos[pOppKnowDeckTroopsNo] = uint8(troopDeckNo)
	if len(tacOnHandixs)+len(troopOnHandixs) > 0 {
		cardixs := append(tacOnHandixs, troopOnHandixs...)
		for i, cardix := range cardixs {
			mpos[pOppKnowHand+i] = uint8(cardix)
		}
	}
}
func countCardsHandDeck(cardixs []int, mpos MPos) (onHandixs []int, deckNo int) {
	for _, cardix := range cardixs {
		if mpos[cardix] == CardPosAll.Deck {
			deckNo = deckNo + 1
		} else if mpos[cardix] == CardPosAll.Hand || mpos[cardix] == CardPosAll.HandLegal {
			onHandixs = append(onHandixs, cardix)
		}
	}
	return onHandixs, deckNo
}
func setScoutReturnBotFields(tacs, troops []int, mpos MPos) {

	if len(tacs) == 1 {
		mpos[pScoutRBotFirstCard] = uint8(tacs[0])
		if len(troops) == 1 {
			mpos[pScoutRBotSecondCard] = uint8(troops[0])
		}
	} else {
		if len(tacs) == 2 {
			mpos[pScoutRBotFirstCard] = uint8(tacs[0])
			mpos[pScoutRBotSecondCard] = uint8(tacs[1])
		} else if len(troops) == 2 {
			mpos[pScoutRBotFirstCard] = uint8(troops[0])
			mpos[pScoutRBotSecondCard] = uint8(troops[1])
		}
	}
	scoutReturnBotClear(mpos, pScoutRBotFirstCard)
	scoutReturnBotClear(mpos, pScoutRBotSecondCard)

}
func scoutReturnBotClear(mpos MPos, ix int) {
	if mpos[ix] != 0 {
		cardPos := mpos[mpos[ix]]
		if !(cardPos == CardPosAll.Deck || cardPos == CardPosAll.Hand || cardPos == CardPosAll.HandLegal) {
			mpos[ix] = 0
		}
	}
}
func setBotCardsPos(pos *bot.Pos) MPos {
	m := NewMPos()
	startix := pCard - 1 //-1 is because cardix is not zero indexed
	for i := pCard; i <= cards.NOTac+cards.NOTroop; i++ {
		m[i] = CardPosAll.Deck
	}
	for _, cardix := range pos.PlayHand.Troops {
		if cardix != 0 {
			m[startix+cardix] = CardPosAll.Hand
		}
	}
	for _, cardix := range pos.PlayHand.Tacs {
		if cardix != 0 {
			m[startix+cardix] = CardPosAll.Hand
		}
	}
	for cardixTxt := range pos.Turn.MovesHand {
		cardix, _ := strconv.Atoi(cardixTxt)
		m[startix+cardix] = CardPosAll.HandLegal
	}
	for flagix, flag := range pos.Flags {
		setCardPosFlag(m, flag.PlayEnvs, flagix, false, startix)
		setCardPosFlag(m, flag.PlayTroops, flagix, false, startix)
		setCardPosFlag(m, flag.OppEnvs, flagix, true, startix)
		setCardPosFlag(m, flag.OppTroops, flagix, true, startix)
	}
	setCardPosDish(m, pos.OppDish.Tacs, true, startix)
	setCardPosDish(m, pos.OppDish.Troops, true, startix)
	setCardPosDish(m, pos.PlayDish.Tacs, false, startix)
	setCardPosDish(m, pos.PlayDish.Troops, false, startix)
	return m
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

// Domain a domain interface.
type Domain interface {
	Name() string
	TxtToValue(string) uint8
	ValueToTxt(uint8) string
	DomValues() []uint8
}

//SimpleDom a standard domain just integer and text.
type SimpleDom struct {
	name      string
	values    []uint8
	txtValues []string
}

// DomValues returns the domain values.
func (dom *SimpleDom) DomValues() []uint8 {
	return dom.values
}

// Name the domain name.
func (dom *SimpleDom) Name() string {
	return dom.name
}

// ValueToTxt translate from value to text
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

// TxtToValue translate from text to value.
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
	//CardPosAll a all card postion domain.
	CardPosAll *CardPosAllSingleton
	//ConePos a all cone postion domain.
	ConePos *ConePosSingleton
)

func init() {
	CardPosAll = newCardPosAllSingleton()
	ConePos = newConePosSingleton()
}

// CardPosAllSingleton all card positions domain.
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
		cp.txtValues[botix] = fmt.Sprintf("Flag%vBot", i+1)
		cp.txtValues[oppix] = fmt.Sprintf("Flag%vOpp", i+1)
	}
	cp.HandLegal = 20
	cp.txtValues[20] = "HandLegal"
	cp.Hand = 21
	cp.txtValues[21] = "Hand"
	cp.Deck = 22
	cp.txtValues[22] = "Deck"

	return cp
}

// BotFlagsValue returns all flag positions 1-9
func (cp *CardPosAllSingleton) BotFlagsValue() []uint8 {
	return []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9}
}

// Flagix return the flag index from position.
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

// CardPosDom a card postion domain.
type CardPosDom struct {
	name   string
	values []uint8
}

// NewCardPosDom create new card position domain.
func NewCardPosDom(values []uint8) (cp *CardPosDom) {
	cp = new(CardPosDom)
	cp.values = values
	return cp
}

// Name returns the domain name.
func (cp *CardPosDom) Name() string {
	return cp.name
}

// TxtToValue translates domain text to value.
func (cp *CardPosDom) TxtToValue(txtValue string) uint8 {
	value := CardPosAll.TxtToValue(txtValue)
	_, ok := domValueIndex(value, cp.values)

	if !ok {
		panic(fmt.Sprintf("Domain text value %v does not exist", txtValue))
	}
	return value
}

// ValueToTxt translates domain value to text.
func (cp *CardPosDom) ValueToTxt(value uint8) string {
	_, ok := domValueIndex(value, cp.values)
	if !ok {
		panic(fmt.Sprintf("Domain value %v does not exist", value))
	}
	return CardPosAll.ValueToTxt(value)
}

// DomValues returns all domain values.
func (cp *CardPosDom) DomValues() []uint8 {
	return cp.values
}

// ConePosSingleton all the cone positions domain.
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

// MPosFld a machine position field.
type MPosFld struct {
	Fld
	MposIx int
}

// MPosCreateFeatureFlds create feature fields.
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
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Scout Return Bot First Card", Domain: noScoutDomain}}
	ix = ix + 1
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Scout Return Bot Second Card", Domain: noScoutDomain}}
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

	// ------------Move part--------------
	ix = ix + 1
	allCardDomain := newCardDomain("All cards", nil)
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Move First Card", Domain: allCardDomain}}
	ix = ix + 1
	firstDestDom, secondDestDom := createMoveDestDoms()
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Move First Card Destination", Domain: firstDestDom}}

	noGuileDomain := newCardDomain("No guile cards", []uint8{uint8(cards.TCScout), uint8(cards.TCTraitor), uint8(cards.TCDeserter), uint8(cards.TCRedeploy)})
	ix = ix + 1
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Move Second Card", Domain: noGuileDomain}}
	ix = ix + 1
	flds[ix] = MPosFld{MposIx: ix + 1, Fld: &DomFld{Name: "Move Second Card Destination", Domain: secondDestDom}}

	return flds
}

func createMoveDestDoms() (firstCardDom, secondCardDom *CardPosDom) {
	moveDests := make([]uint8, 0, 11)
	moveDests = append(moveDests, CardPosAll.DishBot)
	for i := 0; i < 9; i++ {
		moveDests = append(moveDests, CardPosAll.FlagBot[i])
	}
	moveDests = append(moveDests, CardPosAll.DishOpp)
	firstCardDom = NewCardPosDom(moveDests[:10])
	secondCardDom = NewCardPosDom(moveDests)
	return firstCardDom, secondCardDom
}

// CardDomain a card domain a list of cards.
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

// Name returns the name.
func (cd *CardDomain) Name() (name string) {
	return cd.name
}

// TxtToValue returns the card value(index) from text.
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

// ValueToTxt returns the card name from value(index).
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

// DomValues return all the card values.
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
	pPass                = 88 //1850 1 features 1850
)
