package game

var (
	//MoveTypeAll all the move types
	MoveTypeAll MoveTypeAllST
)

func init() {
	MoveTypeAll = newMoveTypeAllST()
}

//MoveType a move type Cone,Hand,Init,Deck ...
//WARNING json of slice of uint8 behaves different that a single value see card or pos
//implementation if that is need.
type MoveType uint8

func (m MoveType) String() string {
	return MoveTypeAll.Names()[int(m)]
}

//IsPause is the move the pause move.
func (m MoveType) IsPause() bool {
	return m == MoveTypeAll.Pause
}

//IsHand returns true if moveType is hand or scout1
func (m MoveType) IsHand() bool {
	return m == MoveTypeAll.Hand || m == MoveTypeAll.Scout1
}

//IsScout returns true if the move is part of the scout move.
//play scout and draw one card, draw second card and draw 3 card.
func (m MoveType) IsScout() bool {
	return m == MoveTypeAll.Scout1 || m == MoveTypeAll.Scout2 || m == MoveTypeAll.Scout3
}

//IsChangePlayer returns true if the next move is by a new player.
func (m MoveType) isChangePlayer() bool {
	return m == MoveTypeAll.Init ||
		m == MoveTypeAll.Deck ||
		m == MoveTypeAll.ScoutReturn
}

//HasNext true if the game continues after this move.
func (m MoveType) HasNext() bool {
	return m != MoveTypeAll.Pause && m != MoveTypeAll.GiveUp
}

//Next returns the move type that follow.
func (m MoveType) Next(mover int) (moveType MoveType, nextMover int) {
	next := [...]MoveType{MoveTypeAll.Hand, MoveTypeAll.Hand, MoveTypeAll.Cone,
		MoveTypeAll.Deck, MoveTypeAll.Scout2, MoveTypeAll.Scout3, MoveTypeAll.ScoutReturn,
		MoveTypeAll.Cone, MoveTypeAll.None, MoveTypeAll.None, MoveTypeAll.Init}
	moveType = next[int(m)]
	if m.isChangePlayer() {
		nextMover = opp(mover)
	} else {
		nextMover = mover
	}
	return moveType, nextMover
}

//MoveTypeAllST All the move types.
type MoveTypeAllST struct {
	Init        MoveType
	Cone        MoveType
	Deck        MoveType
	Hand        MoveType
	Scout1      MoveType
	Scout2      MoveType
	Scout3      MoveType
	ScoutReturn MoveType
	GiveUp      MoveType
	Pause       MoveType
	None        MoveType
}

func newMoveTypeAllST() (m MoveTypeAllST) {
	m.Init = 0
	m.Cone = 1
	m.Deck = 2
	m.Hand = 3
	m.Scout1 = 4
	m.Scout2 = 5
	m.Scout3 = 6
	m.ScoutReturn = 7
	m.GiveUp = 8
	m.Pause = 9
	m.None = 10
	return m
}

// Names returns all the move types name.
func (m MoveTypeAllST) Names() []string {
	return []string{"Init", "Cone", "Deck", "Hand",
		"Scout1", "Scout2", "Scout3", "Scout-Return", "Give-Up", "Pause", "None"}
}

// All returns all the move types.
func (m MoveTypeAllST) All() []MoveType {
	return []MoveType{0, 1, 2, 3, 4, 5, 6, 7, 8}

}
