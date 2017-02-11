package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	bat "github.com/rezder/go-battleline/battleline"
	pub "github.com/rezder/go-battleline/battserver/publist"
	"github.com/rezder/go-battleline/battserver/tables"
)

func unmarshalMoveJSON(data []byte) (mv *pub.MoveView, err error) {
	//TODO Maybe move all tables' moves to public, and this to public as metode of MoveView
	mv = new(pub.MoveView)
	mv.Turn = new(pub.Turn)
	var raw moveViewRaw
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return mv, err
	}
	mv.Mover = raw.Mover
	mv.MoveCardix = raw.MoveCardix
	mv.DeltCardix = raw.DeltCardix
	mv.MyTurn = raw.MyTurn
	mv.State = raw.State
	mv.MovesPass = raw.MovesPass
	mv.Move, err = moveUnmarshal(raw.Move)
	if err != nil {
		return mv, err
	}
	if raw.Moves != nil {
		mv.Moves, err = sliceMoveUnmarshal(raw.Moves)
		if err != nil {
			return mv, err
		}
	}

	if raw.MovesHand != nil {
		mv.MovesHand = make(map[string][]bat.Move)
		var sms []bat.Move
		for k, v := range raw.MovesHand {
			sms, err = sliceMoveUnmarshal(v)
			if err != nil {
				return mv, err
			}
			mv.MovesHand[k] = sms
		}
	}
	return mv, err
}
func sliceMoveUnmarshal(data []json.RawMessage) (moves []bat.Move, err error) {
	moves = make([]bat.Move, len(data))
	var move bat.Move
	for i, v := range data {
		move, err = moveUnmarshal(v)
		if err != nil {
			return moves, err
		}
		moves[i] = move
	}
	return moves, err
}
func moveUnmarshal(data []byte) (move bat.Move, err error) {
	var raw moveRaw
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return move, err
	}
	switch raw.JsonType {
	case "MoveInit":
		var moveInit tables.MoveInit
		err = json.Unmarshal(data, &moveInit)
		if err == nil {
			move = moveInit
		}
	case "MoveInitPos":
		var moveInitPos tables.MoveInitPos
		err = json.Unmarshal(data, &moveInitPos)
		if err == nil {
			move = moveInitPos
		}
	case "MoveCardFlag":
		var moveCardFlag bat.MoveCardFlag
		err = json.Unmarshal(data, &moveCardFlag)
		if err == nil {
			move = moveCardFlag
		}
	case "MoveDeck":
		var moveDeck bat.MoveDeck
		err = json.Unmarshal(data, &moveDeck)
		if err == nil {
			move = moveDeck
		}
	case "MoveClaim":
		var moveClaim bat.MoveClaim
		err = json.Unmarshal(data, &moveClaim)
		if err == nil {
			move = moveClaim
		}
	case "MoveClaimView":
		var moveClaimView tables.MoveClaimView
		err = json.Unmarshal(data, &moveClaimView)
		if err == nil {
			move = moveClaimView
		}
	case "MoveDeserter":
		var moveDeserter bat.MoveDeserter
		err = json.Unmarshal(data, &moveDeserter)
		if err == nil {
			move = moveDeserter
		}
	case "MoveDeserterView":
		var moveDeserterView tables.MoveDeserterView
		err = json.Unmarshal(data, &moveDeserterView)
		if err == nil {
			move = moveDeserterView
		}
	case "MoveScoutReturn":
		var moveSR bat.MoveScoutReturn
		err = json.Unmarshal(data, &moveSR)
		if err == nil {
			move = moveSR
		}
	case "MoveScoutReturnView":
		var moveSRW tables.MoveScoutReturnView
		err = json.Unmarshal(data, &moveSRW)
		if err == nil {
			move = moveSRW
		}
	case "MoveTraitor":
		var moveTraitor bat.MoveTraitor
		err = json.Unmarshal(data, &moveTraitor)
		if err == nil {
			move = moveTraitor
		}
	case "MoveRedeploy":
		var moveRedeploy bat.MoveRedeploy
		err = json.Unmarshal(data, &moveRedeploy)
		if err == nil {
			move = moveRedeploy
		}
	case "MoveRedeployView":
		var moveRW tables.MoveRedeployView
		err = json.Unmarshal(data, &moveRW)
		if err == nil {
			move = moveRW
		}
	case "MovePass":
		var movePass tables.MovePass
		err = json.Unmarshal(data, &movePass)
		if err == nil {
			move = movePass
		}
	case "MoveQuit":
		var moveQuit tables.MoveQuit
		err = json.Unmarshal(data, &moveQuit)
		if err == nil {
			move = moveQuit
		}
	case "MoveSave":
		var moveSave tables.MoveSave
		err = json.Unmarshal(data, &moveSave)
		if err == nil {
			move = moveSave
		}
	default:
		err = errors.New("Missing json type implementation for move: " + raw.JsonType)
	}
	return move, err
}

type moveViewRaw struct {
	Move       json.RawMessage
	MoveCardix int
	//DeltCardix is the return card from deck moves, zero when not used.
	DeltCardix int
	State      int
	Moves      []json.RawMessage
	MovesHand  map[string][]json.RawMessage
	Mover      bool
	MyTurn     bool
	MovesPass  bool
}

type moveRaw struct {
	JsonType string
}
