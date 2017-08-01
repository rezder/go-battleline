package conv

import (
	"encoding/gob"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/battarchiver/battdb"
	bold "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/v2/db/dbhist"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
	"github.com/rezder/go-error/log"
	"os"
	"time"
)

//Game convert a old game game history.
func Game(oldGame *bold.Game, gameTime time.Time) (hist *game.Hist) {
	moves := make([]*game.Move, 0, len(oldGame.Moves)+2)
	hist = &game.Hist{
		Moves:     moves,
		PlayerIDs: oldGame.PlayerIds,
		Time:      gameTime,
	}
	oldGame.GameMoveLoop(func(
		gameMoveix int,
		pos *bold.GamePos,
		moveCardix int,
		dealtix int,
		moveix int,
		move bold.Move,
		isGiveUp bool,
		isPass bool,
		claimFailExs [9][]int,
	) {
		if gameMoveix == 0 {
			hist.AddMove(createInitMove(pos))
		}
		if isPass {
			passMove := game.NewMove(pos.Player, game.MoveTypeAll.Hand)
			hist.AddMove(passMove)
		} else if isGiveUp {
			hist.AddMove(game.NewMove(pos.Player, game.MoveTypeAll.GiveUp))
		} else {
			hist.AddMove(Move(move, pos.Player, moveCardix, dealtix, pos.Turn.State, claimFailExs))
		}
	})
	if oldGame.Pos.State != bold.TURNFinish && oldGame.Pos.State != bold.TURNQuit {
		hist.AddMove(game.NewMove(oldGame.Pos.Player, game.MoveTypeAll.Pause))
	}
	return hist
}
func opp(player int) int {
	o := player + 1
	if o > 1 {
		o = 0
	}
	return o
}
func createInitMove(oldPos *bold.GamePos) (initMove *game.Move) {
	mover := opp(oldPos.Player)
	initMove = game.NewMove(mover, game.MoveTypeAll.Init)
	initMove.Moves = make([]*game.BoardPieceMove, 0, 14)
	for i, moverTroopix := range oldPos.Hands[mover].Troops {
		oppTroopix := oldPos.Hands[opp(mover)].Troops[i]
		move0 := game.BoardPieceMove{
			BoardPiece: game.BoardPieceAll.Card,
			Index:      oppTroopix,
			OldPos:     uint8(pos.CardAll.DeckTroop),
			NewPos:     uint8(pos.CardAll.Players[opp(mover)].Hand),
		}
		initMove.Moves = append(initMove.Moves, &move0)
		move1 := game.BoardPieceMove{
			BoardPiece: game.BoardPieceAll.Card,
			Index:      moverTroopix,
			OldPos:     uint8(pos.CardAll.DeckTroop),
			NewPos:     uint8(pos.CardAll.Players[mover].Hand),
		}
		initMove.Moves = append(initMove.Moves, &move1)
	}
	return initMove
}

//Move converts old move to new move.
func Move(oldMove bold.Move, mover, cardix, dealtix, turnState int, claimFailExs [9][]int) (newMove *game.Move) {
	switch cardMove := oldMove.(type) {
	case bold.MoveCardFlag:
		newMove = game.CreateMoveHand(cardix, cardMove.Flagix, mover)
	case bold.MoveDeserter:
		dishPBMove := game.CreateBPMoveDish(cardix, mover, pos.CardAll.Players[mover].Hand)
		oldPos := pos.CardAll.Players[opp(mover)].Flags[cardMove.Flag]
		newMove = game.CreateMoveDeserter(oldPos, cardMove.Card, mover, dishPBMove)
	case bold.MoveRedeploy:
		dishPBMove := game.CreateBPMoveDish(cardix, mover, pos.CardAll.Players[mover].Hand)
		oldPos := pos.CardAll.Players[mover].Flags[cardMove.OutFlag]
		var newPos pos.Card
		if cardMove.InFlag == bold.REDeployDishix {
			newPos = pos.CardAll.Players[mover].Dish
		} else {
			newPos = pos.CardAll.Players[mover].Flags[cardMove.InFlag]
		}
		newMove = game.CreateMoveDouble(oldPos, newPos, cardMove.OutCard, mover, dishPBMove)
	case bold.MoveTraitor:
		dishPBMove := game.CreateBPMoveDish(cardix, mover, pos.CardAll.Players[mover].Hand)
		oldPos := pos.CardAll.Players[opp(mover)].Flags[cardMove.OutFlag]
		newPos := pos.CardAll.Players[mover].Flags[cardMove.InFlag]
		newMove = game.CreateMoveDouble(oldPos, newPos, cardMove.OutCard, mover, dishPBMove)
	case bold.MoveScoutReturn:
		newMove = game.CreateMoveScoutReturn(cardMove.Tac, cardMove.Troop, mover)
	case bold.MoveDeck:
		deckMove := createMoveDeck(dealtix, mover)
		switch turnState {
		case bold.TURNDeck:
			newMove = deckMove
		case bold.TURNHand:
			dishBPMove := game.CreateBPMoveDish(cardix, mover, pos.CardAll.Players[mover].Hand)
			newMove = game.CreateMoveScout(dishBPMove, deckMove.Moves[0], mover)
		case bold.TURNScout1:
			newMove = deckMove
			newMove.MoveType = game.MoveTypeAll.Scout3
		case bold.TURNScout2:
			newMove = deckMove
			newMove.MoveType = game.MoveTypeAll.Scout2
		}
	case bold.MoveClaim:
		flagixs := cardMove.Flags
		if len(cardMove.Flags) > 0 {
			flagixs = make([]int, 0, len(cardMove.Flags))
			for _, flagix := range cardMove.Flags {
				if len(claimFailExs[flagix]) == 0 {
					flagixs = append(flagixs, flagix)
				}
			}
		}
		newMove = game.CreateMoveCone(flagixs, mover)
	}
	return newMove
}
func createMoveDeck(cardix, mover int) (move *game.Move) {
	move = game.NewMove(mover, game.MoveTypeAll.Deck)
	oldPos := pos.CardAll.DeckTroop
	if card.Move(cardix).IsTac() {
		oldPos = pos.CardAll.DeckTac
	}
	move.Moves = append(move.Moves, &game.BoardPieceMove{
		BoardPiece: game.BoardPieceAll.Card,
		Index:      cardix,
		OldPos:     uint8(oldPos),
		NewPos:     uint8(pos.CardAll.Players[mover].Hand),
	})
	return move
}
func loadOldGame(filePath string) (game *bold.Game, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		err = errors.Wrapf(err, "Open file: %v failed", filePath)
		return game, err
	}
	game, err = bold.Load(file)
	if err != nil {
		err = errors.Wrapf(err, "Load game file: %v failed", filePath)
		_ = file.Close()
		return game, err
	}
	err = file.Close()
	if err != nil {
		err = errors.Wrapf(err, "Closing file: %v failed", filePath)
	}
	return game, err
}

//GameFile convert a battleline game file to Hist file
func GameFile(src, dest string) (err error) {
	oldGame, err := loadOldGame(src)
	if err != nil {
		err = errors.Wrapf(err, "Error loading old game from file:%v", src)
		return err
	}
	hist := Game(oldGame, time.Now())
	err = saveHist(dest, hist)
	return err
}
func saveHist(filePath string, hist *game.Hist) (err error) {
	destFile, err := os.Create(filePath)
	if err != nil {
		err = errors.Wrapf(err, "Opening file: %v failed", filePath)
		return err
	}
	err = gob.NewEncoder(destFile).Encode(hist)
	if err != nil {
		_ = destFile.Close()
		err = errors.Wrapf(err, "Encoding file: %v failed", filePath)
		return err
	}
	err = destFile.Close()
	if err != nil {
		err = errors.Wrapf(err, "Closing file: %v failed", filePath)
		return err
	}
	return err
}

//DbFile covert a old game bolt database file.
func DbFile(src, dst string) (err error) {
	oldDb, err := bolt.Open(src, 0600, nil)
	if err != nil {
		err = errors.Wrapf(err, "Open data base file %v failed", src)
		return err
	}
	defer func() {
		if cerr := oldDb.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	gameDb := battdb.New(battdb.KeyPlayersTime, oldDb, 1000)
	err = gameDb.Init()
	if err != nil {
		err = errors.Wrapf(err, "Init database %v failed", src)
		return err
	}

	newDb, err := bolt.Open(dst, 0600, nil)
	if err != nil {
		err = errors.Wrapf(err, "Open data base file %v failed", dst)
		return err
	}
	defer func() {
		cerr := newDb.Close()
		if cerr != nil {
			if err == nil {
				err = cerr
			}
		} else if rerr := os.Remove(dst); rerr != nil && err == nil {
			err = rerr
		}
	}()

	dbHist := dbhist.New(dbhist.KeyPlayersTime, newDb, 1000)
	err = dbHist.Init()
	if err != nil {
		err = errors.Wrapf(err, "Init database %v failed", dst)
		return err
	}
	var nextKey []byte
	var games []*bold.Game
	for {
		games, nextKey, err = gameDb.SearchLoop(nil, nextKey)
		if err != nil {
			err = errors.Wrapf(err, "Search game database: %v failed", src)
			return err
		}
		log.Printf(log.Min, "Loaded %v games", len(games))
		hists := make([]*game.Hist, 0, len(games))
		for _, game := range games {
			hists = append(hists, Game(game, time.Now()))
		}
		err = dbHist.Puts(hists)
		if err != nil {
			err = errors.Wrapf(err, "Puts hist database: %v failed", dst)
			return err
		}
		if len(nextKey) == 0 {
			break
		}
	}
	return err
}
