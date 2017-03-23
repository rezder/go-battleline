package machine

import (
	"bytes"
	"encoding/gob"

	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
)

// Meta holds meta game information.
type Meta struct {
	Players   [2]*MetaPlayer
	Winerix   uint8
	Starterix uint8
	GiveUp    bool
	NoMoves   uint8
}

func NewMeta(playerids [2]int, starterix int) (m *Meta) {
	m = new(Meta)
	m.Starterix = uint8(starterix)
	m.Players = [2]*MetaPlayer{NewMetaPlayer(playerids[0]), NewMetaPlayer(playerids[0])}
	return m
}
func gobInitMeta() (m *Meta) {
	m = new(Meta)
	m.Players = [2]*MetaPlayer{gobInitMetaPlayer(), gobInitMetaPlayer()}
	return m
}
func (m *Meta) SetWinner(ix int) {
	m.Winerix = uint8(ix)
}
func (m *Meta) SetNoMoves(ix int) {
	m.NoMoves = uint8(ix)
}
func (m *Meta) AddLastPosInfo(pos *bat.GamePos) {
	troops := [2][]int{make([]int, 0, 37), make([]int, 0, 37)}
	for i, player := range m.Players {
		for _, tacix := range player.PlayedTacixs {
			player.TaticCardixs = append(player.TaticCardixs, tacix)
		}
		for _, cardix := range pos.Hands[i].Tacs {
			if cards.IsGuile(cardix) {
				player.TaticCardixs = append(player.TaticCardixs, cardix)
			}
		}
		for _, cardix := range pos.Hands[i].Troops {
			troops[i] = append(troops[i], cardix)
		}
		for _, cardix := range pos.Dishs[i].Troops {
			troops[i] = append(troops[i], cardix)
		}
		for _, f := range pos.Flags {
			for _, cardix := range f.Players[i].Troops {
				if cards.IsTroop(cardix) {
					troops[i] = append(troops[i], cardix)
				}
			}
		}
		if len(troops[i]) > 0 {
			sum := 0
			for _, troopix := range troops[i] {
				troop, _ := cards.DrTroop(troopix)
				sum = sum + troop.Value()
			}
			if player.traitorCardix != 0 {
				troop, _ := cards.DrTroop(player.traitorCardix)
				sum = sum - troop.Value()

			} else if m.Players[opponent(i)].traitorCardix != 0 {
				troop, _ := cards.DrTroop(m.Players[opponent(i)].traitorCardix)
				sum = sum + troop.Value()
			}
			player.TroopsAvgStrenght = float32(sum) / float32(len(troops[i]))
		}
	}

	//Avg sum on hand on flags on dish plus and minus traitor card
}

// MetaPlayers holds meta game information per player.
type MetaPlayer struct {
	PlayerId          int
	TaticCardixs      []int
	PlayedTacixs      []int
	TroopsAvgStrenght float32
	NoStdMoves        uint8
	NoSpecialMoves    uint8
	traitorCardix     int
}

func NewMetaPlayer(id int) (m *MetaPlayer) {
	m = gobInitMetaPlayer()
	m.PlayerId = id
	return m
}
func gobInitMetaPlayer() (m *MetaPlayer) {
	m = new(MetaPlayer)
	m.TaticCardixs = make([]int, 0, 4)
	m.PlayedTacixs = make([]int, 0, 4)
	return m
}

func (m *MetaPlayer) AddHandMove(move bat.Move) {
	switch cardMove := move.(type) {
	case bat.MoveCardFlag:
		m.NoStdMoves = m.NoStdMoves + 1
	case bat.MoveDeserter:
		m.NoSpecialMoves = m.NoSpecialMoves + 1
	case bat.MoveRedeploy:
		m.NoSpecialMoves = m.NoSpecialMoves + 1
	case bat.MoveDeck: //Scout
		m.NoSpecialMoves = m.NoSpecialMoves + 1
	case bat.MoveTraitor:
		m.NoSpecialMoves = m.NoSpecialMoves + 1
		m.traitorCardix = cardMove.OutCard
	}
}

// MetaEncode encodes Meta to bytes.
func MetaEncode(m *Meta) (value []byte, err error) {
	var mBuf bytes.Buffer
	encoder := gob.NewEncoder(&mBuf)
	err = encoder.Encode(m)
	if err != nil {
		return value, err
	}
	value = mBuf.Bytes()
	return value, err
}

// MetaDecode decodes meta from bytes
func MetaDecode(value []byte) (meta *Meta, err error) {
	buf := bytes.NewBuffer(value)
	decoder := gob.NewDecoder(buf)
	m := *new(Meta)
	err = decoder.Decode(&m)
	if err != nil {
		return meta, err
	}
	meta = &m
	return meta, err
}
