package game

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGame(t *testing.T) {
	game := NewGame()
	game.Start([2]int{1, 2}, 0)
	testCheckCardsOnhand([2]int{7, 7}, game.Pos.CardPos, t)
	moves := game.Pos.CalcMoves()
	noMoves := len(moves)
	noExpMoves := 9 * 7
	if noMoves != noExpMoves {
		t.Errorf("No. of moves should be %v after init it is: %v", noExpMoves, noMoves)
	}
	winner, _ := game.Move(moves[0])
	no := 1
	for winner == NoPlayer {
		moves = game.Pos.CalcMoves()
		winner, _ = game.Move(testMove(moves))
		no++
		if no == 3 {
			pauseMove := NewMove(game.Pos.LastMover, MoveTypeAll.Pause)
			game.Move(pauseMove)
			game.Resume()
			moves = game.Pos.CalcMoves()
			game.Move(moves[0]) //DeckTac
		}
	}
	lastPos := *game.Pos
	for game.ScrollBackward() {
	}
	ok := game.Resume()
	if ok {
		t.Error("Game should be finished and resume should return false")
	}
	if !game.Pos.IsEqual(&lastPos) {
		t.Errorf("Game postion should be the same!\n Pos:\n%v\nPos Scroll:\n%v\n", lastPos, game.Pos)
	}
	gameHist := game.Hist.Copy()
	game2 := NewGame()
	game2.LoadHist(gameHist)
	ok = game2.Resume()
	if ok {
		t.Error("Game should be finished and resume should return false")
	}
	if !game2.Pos.IsEqual(&lastPos) {
		t.Errorf("Game postion should be the same!\n Pos:\n%v\nPos Load hist:\n%v\n", lastPos, game.Pos)
	}
	testGob(gameHist, t)
	testJSON(gameHist, t)
}
func testCheckCardsOnhand(expNos [2]int, cardPos [71]pos.Card, t *testing.T) {
	posCards := NewPosCards(cardPos)
	for i, expNo := range expNos {
		noHand := len(posCards.Cards(pos.CardAll.Players[i].Hand))
		if noHand != expNo {
			t.Errorf("Player %v no. cards: %v deviates from expected: %v", i, noHand, expNo)
		}
	}
}
func testGob(gameHist *Hist, t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(gameHist)
	if err != nil {
		t.Errorf("Error encoding: %v", err)
		return
	}

	var loadGameHist Hist
	decoder := gob.NewDecoder(buf)

	// Decoding the serialized data
	err = decoder.Decode(&loadGameHist)
	if err != nil {
		t.Errorf("Error decoding: %v", err)
	} else {
		if !gameHist.IsEqual(&loadGameHist) {
			t.Error("Loaded history have been changed")
		}
	}
}
func testJSON(gameHist *Hist, t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(gameHist)
	if err != nil {
		t.Errorf("Error encoding: %v", err)
		return
	}

	var loadGameHist Hist
	decoder := json.NewDecoder(buf)

	// Decoding the serialized data
	err = decoder.Decode(&loadGameHist)
	if err != nil {
		t.Errorf("Error decoding: %v", err)
	} else {
		if !gameHist.IsEqual(&loadGameHist) {
			t.Error("Loaded history have been changed")
		}
	}
}
func testMove(moves []*Move) (move *Move) {
	moveix := 0
	switch moves[0].MoveType {
	case MoveTypeAll.Cone:
		moveix = 1
	case MoveTypeAll.Deck:
		if len(moves) == 2 {
			moveix = 1
		}
	case MoveTypeAll.Hand:
		moveix = rand.Intn(len(moves) - 1)
	}
	return moves[moveix]
}
func testFindGameFiles(dirName string, t *testing.T) (posFileNameSet map[string]bool, newGameFileNames []string) {
	posFileNameSet = make(map[string]bool)
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		t.Errorf("Error listing directory: %v. Error: %v", dirName, err)
		return
	}

	for _, fileInfo := range files {
		if strings.Contains(fileInfo.Name(), "pos") {
			posFileNameSet[fileInfo.Name()] = true
		}
	}
	for _, fileInfo := range files {
		if !strings.Contains(fileInfo.Name(), "pos") && !fileInfo.IsDir() {
			if !posFileNameSet["pos"+fileInfo.Name()] {
				newGameFileNames = append(newGameFileNames, fileInfo.Name())
			}
		}
	}
	return posFileNameSet, newGameFileNames
}
func TestSavedGame(t *testing.T) {
	dirName := filepath.Join(testDir, "games")
	posFileNameSet, newGameFileNames := testFindGameFiles(dirName, t)
	posFileNameSet = testCreatePos(t, newGameFileNames, posFileNameSet, dirName)
	for fileName := range posFileNameSet {
		game, testPoss := testLoadPosGames(t, fileName, dirName)
		if testPoss != nil {
			posix := 0
			isNext := true
			winner := NoPlayer
			for isNext {
				winner, isNext = game.ScrollForward() //Init Move
				if isNext || winner != NoPlayer {
					var moves []*Move
					if winner == NoPlayer {
						moves = game.Pos.CalcMoves()
					}
					oldPos := testPoss[posix].GamePos
					oldMoves := testPoss[posix].Moves
					if !game.Pos.IsEqual(oldPos) {
						t.Errorf("File: %v postion index: %v deviate, Saved position: %v\n Postion: %v",
							fileName, posix, oldPos, game.Pos)
					} else {
						if len(moves) != len(oldMoves) {
							t.Errorf("File: %v moves from postion index: %v deviate in lenght, oldMoves: %v\n moves: %v",
								fileName, posix, oldMoves, moves)
						} else {
							for i, move := range moves {
								if !move.IsEqual(oldMoves[i]) {
									t.Errorf("File: %v moves from postion index: %v deviate move index: %v, oldMove: %v\n move: %v",
										fileName, posix, i, oldMoves, moves)
								}
							}
						}
					}
					posix = posix + 1
				}
			}
		}
	}
}
func testCreatePos(
	t *testing.T,
	newGameNames []string,
	posFileNameSet map[string]bool,
	dirName string,
) map[string]bool {
	for _, fileName := range newGameNames {
		possName := "pos" + fileName
		possPath := filepath.Join(dirName, "pos"+fileName)
		game := testLoadGame(t, fileName, dirName)
		if game != nil {
			testPoss := make([]*TestPos, 0, len(game.Hist.Moves))
			isNext := true
			winner := NoPlayer
			for isNext {
				winner, isNext = game.ScrollForward()
				if isNext || winner != NoPlayer {
					var moves []*Move
					if winner == NoPlayer {
						moves = game.Pos.CalcMoves()
					}
					gamePos := *game.Pos
					testPoss = append(testPoss, &TestPos{
						GamePos: &gamePos,
						Moves:   moves,
					})
				}
			}
			posFile, err := os.Create(possPath)
			if err != nil {
				t.Errorf("Create file error. File :%v Error: %v", possPath, err.Error())
				break
			}
			err = gob.NewEncoder(posFile).Encode(testPoss)
			if err != nil {
				t.Errorf("Encoding postion error. File :%v Error: %v", possPath, err.Error())
				break
			}
			posFileNameSet[possName] = true
		}
	}
	return posFileNameSet
}
func testLoadGame(t *testing.T, fileName, dirName string) (game *Game) {
	filePath := filepath.Join(dirName, fileName)
	var hist Hist
	err := testLoadFile(filePath, &hist)
	if err != nil {
		t.Errorf("Error decoding file: %v Error: %v", filePath, err)
		return game
	}
	game = NewGame()
	game.LoadHist(&hist)
	return game
}
func testLoadPoss(t *testing.T, fileName, dirName string) (poss TestPoss) {
	filePath := filepath.Join(dirName, fileName)
	var loadPoss TestPoss
	err := testLoadFile(filePath, &loadPoss)
	if err != nil {
		t.Errorf("Error decoding file: %v Error: %v", filePath, err)
		return
	}
	return loadPoss
}

func testLoadPosGames(t *testing.T, fileName, dirName string) (game *Game, testPoss TestPoss) {
	game = testLoadGame(t, fileName[3:], dirName)
	if game != nil {
		testPoss = testLoadPoss(t, fileName, dirName)
	}
	return game, testPoss
}

type TestPoss []TestPos
type TestPos struct {
	GamePos *Pos
	Moves   []*Move
}

func TestCalcMoves(t *testing.T) {
	gamePos := NewPos()
	gamePos.CardPos = [71]pos.Card{0, 19, 22, 18, 9, 12, 18, 8, 18, 13, 13, 7, 1, 21, 14, 2, 21, 16, 21, 22, 22, 19, 1, 21, 9, 12, 11, 8, 0, 3, 15, 0, 21, 11, 11, 22, 6, 11, 5, 5, 5, 19, 10, 22, 2, 2, 6, 16, 1, 3, 15, 7, 17, 17, 9, 12, 6, 16, 4, 4, 4, 10, 23, 10, 22, 1, 11, 21, 21, 22, 13}
	gamePos.ConePos = [10]pos.Cone{0, 0, 2, 2, 1, 1, 2, 0, 2, 1}
	gamePos.LastMoveType = MoveTypeAll.Cone
	gamePos.LastMoveIx = 135
	gamePos.LastMover = 1
	moves := gamePos.CalcMoves()
	exp := 7
	if len(moves) != exp {
		t.Log(moves)
		t.Errorf("No. calculated moves: %v deviates form expected moves:%v", len(moves), exp)
	}
	gamePos = NewPos()
	gamePos.CardPos = [71]pos.Card{0, 9, 21, 5, 0, 21, 22, 8, 22, 8, 6, 13, 21, 13, 22, 7, 7, 7, 2, 15, 15, 4, 1, 1, 1, 21, 22, 2, 21, 16, 6, 11, 11, 22, 0, 11, 0, 0, 14, 14, 14, 9, 19, 19, 19, 21, 18, 2, 0, 16, 0, 9, 22, 5, 20, 12, 12, 12, 0, 3, 3, 23, 23, 20, 22, 23, 23, 23, 23, 21, 23}
	gamePos.ConePos = [10]pos.Cone{0, 1, 2, 0, 2, 0, 0, 0, 0, 2}
	gamePos.LastMoveType = MoveTypeAll.Deck
	gamePos.LastMoveIx = 103
	gamePos.LastMover = 0
	moves = gamePos.CalcMoves()
	exp = 30
	if len(moves) != exp {
		t.Log(moves)
		t.Errorf("No. calculated moves: %v deviates form expected moves:%v", len(moves), exp)
	}
	gamePos = NewPos()
	gamePos.CardPos = [71]pos.Card{0, 22, 18, 14, 9, 13, 12, 6, 11, 7, 7, 21, 22, 0, 22, 3, 2, 4, 15, 8, 13, 0, 18, 0, 16, 0, 15, 0, 17, 8, 0, 21, 15, 17, 0, 21, 1, 5, 6, 12, 0, 4, 5, 12, 21, 22, 3, 11, 17, 9, 11, 22, 0, 21, 22, 21, 18, 4, 6, 8, 22, 23, 23, 23, 21, 23, 23, 23, 23, 23, 23}
	gamePos.ConePos = [10]pos.Cone{0, 0, 0, 0, 0, 0, 0, 0, 1, 0}
	gamePos.LastMoveType = MoveTypeAll.Deck
	gamePos.LastMoveIx = 92
	gamePos.LastMover = 0
	moves = gamePos.CalcMoves()
	exp = 16
	if len(moves) != exp {
		t.Log(moves)
		t.Errorf("No. calculated moves: %v deviates form expected moves:%v", len(moves), exp)
	}
}
func TestPass(t *testing.T) {
	gamePos := NewPos()
	gamePos.CardPos = [71]pos.Card{0, 11, 9, 12, 12, 12, 18, 17, 17, 17, 14, 11, 19, 21, 22, 8, 13, 1, 13, 0, 16, 1, 9, 1, 2, 19, 2, 15, 5, 22, 0, 11, 19, 21, 21, 3, 3, 3, 16, 4, 14, 6, 6, 6, 22, 8, 18, 15, 5, 4, 21, 0, 9, 22, 21, 8, 18, 15, 7, 7, 14, 20, 21, 23, 23, 22, 21, 22, 23, 2, 22}
	gamePos.ConePos = [10]pos.Cone{0, 2, 1, 1, 2, 2, 1, 2, 0, 1}
	gamePos.LastMoveType = MoveTypeAll.Cone
	gamePos.LastMoveIx = 124
	gamePos.LastMover = 0
	moves := gamePos.CalcMoves()
	if len(moves[len(moves)-1].Moves) != 0 {
		t.Error("Pass move missing")
	}
	gamePos.AddMove(moves[len(moves)-1])
	moves = gamePos.CalcMoves()
	mover, moveType := moves[0].GetMoverAndType()
	if mover == gamePos.LastMover {
		t.Error("Pass move failed after pass the mover should change as deck have no moves.")
	}
	if moveType != MoveTypeAll.Cone {
		t.Errorf("Pass move failed after pass the move type should be %v.", MoveTypeAll.Cone)
	}
}
func TestView(t *testing.T) {
	gamePos := NewPos()
	gamePos.CardPos = [71]pos.Card{0, 19, 22, 18, 9, 12, 18, 8, 18, 13, 13, 7, 1, 21, 14, 2, 21, 16, 21, 22, 22, 19, 1, 21, 9, 12, 11, 8, 0, 3, 15, 0, 21, 11, 11, 22, 6, 11, 5, 5, 5, 19, 10, 22, 2, 2, 6, 16, 1, 3, 15, 7, 17, 17, 9, 12, 6, 16, 4, 4, 4, 10, 23, 10, 22, 1, 11, 21, 21, 22, 13}
	gamePos.ConePos = [10]pos.Cone{0, 0, 2, 2, 1, 1, 2, 0, 2, 1}
	gamePos.LastMoveType = MoveTypeAll.Cone
	gamePos.LastMoveIx = 136
	gamePos.CardPos[card.TCScout] = pos.CardAll.Players[1].Dish
	gamePos.CardPos[31] = pos.CardAll.Players[1].Hand
	gamePos.LastMover = 0
	gamePos.PlayerReturned = 1
	testViews([2]card.Move{62, 1}, gamePos, t)
	testViews([2]card.Move{62, 2}, gamePos, t)
	testViews([2]card.Move{62, 13}, gamePos, t)
	testViews([2]card.Move{62, 28}, gamePos, t)
}
func testViews(returned [2]card.Move, gamePos *Pos, t *testing.T) {
	gamePos.CardsReturned = returned
	returner := gamePos.PlayerReturned
	posCards := NewPosCards(gamePos.CardPos)
	hand0Troops, hand0Morales, hand0Guiles, hand0Envs := posCards.SortedCards(pos.CardAll.Players[0].Hand)
	hand1Troops, hand1Morales, hand1Guiles, hand1Envs := posCards.SortedCards(pos.CardAll.Players[1].Hand)
	handNoTroops := [2]int{len(hand0Troops), len(hand1Troops)}
	handNoTacs := [2]int{len(hand0Morales) + len(hand0Guiles) + len(hand0Envs), len(hand1Morales) + len(hand1Guiles) + len(hand1Envs)}

	t.Logf("Returned postions: %v,%v", gamePos.CardPos[returned[0]], gamePos.CardPos[returned[1]])
	for _, v := range ViewAll.All() {
		var noExpTacs [2]int
		var noExpTroops [2]int
		var expReturned [2]card.Move
		switch v {
		case ViewAll.God:
			expReturned = returned
		case ViewAll.Spectator:
			noExpTroops = handNoTroops
			noExpTacs = handNoTacs
			for i, cardMove := range returned {
				if gamePos.CardPos[cardMove].IsInDeck() {
					expReturned[i] = getBack(cardMove)
				} else {
					expReturned[i] = 0
				}
			}
		default:
			viewer := v.Playerix()
			noExpTroops = handNoTroops
			noExpTacs = handNoTacs
			noExpTroops[v.Playerix()] = 0
			noExpTacs[v.Playerix()] = 0
			for i, cardMove := range returned {
				if returner == viewer {
					expReturned[i] = cardMove
				} else {
					if gamePos.CardPos[cardMove].IsInDeck() {
						expReturned[i] = getBack(cardMove)
					} else if gamePos.CardPos[cardMove].IsOnHand() && gamePos.CardPos[cardMove].Player() == viewer {
						expReturned[i] = cardMove
					} else {
						expReturned[i] = 0
					}
				}
			}
			if returner == v.Playerix() {
				for _, cardMove := range returned {
					if gamePos.CardPos[cardMove].IsOnHand() && gamePos.CardPos[cardMove].Player() != viewer {
						if cardMove.IsTac() {
							noExpTacs[opp(viewer)] = noExpTacs[opp(viewer)] - 1
						} else {
							noExpTroops[opp(viewer)] = noExpTroops[opp(viewer)] - 1
						}
					}
				}
			}
		}
		testCheckView(NewViewPos(gamePos, v, NoPlayer), noExpTacs, noExpTroops, expReturned, t)
	}
}
func getBack(cardMove card.Move) card.Move {
	if cardMove.IsTac() {
		return card.BACKTac
	}
	return card.BACKTroop
}

func testCheckView(viewPos *ViewPos, noTacs, noTroops [2]int, returned [2]card.Move, t *testing.T) {
	if noTacs != viewPos.NoTacs || noTroops != viewPos.NoTroops || returned != viewPos.CardsReturned {
		t.Errorf("View:%v deviates expected %v,%v,%v got %v,%v,%v", viewPos.View, noTroops, noTacs, returned, viewPos.NoTroops, viewPos.NoTacs, viewPos.CardsReturned)
	}
}
