package battleline

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"testing"
)

func TestGameSave(t *testing.T) {
	fileName := "test/save.gob"
	game := New(1, [2]int{1, 2})
	game.Start(1)
	file, err := os.Create(fileName)
	//f,err:=os.Open(fileName)
	GobRegistor()
	if err == nil {
		err = Save(game, file)
		file.Close()
		if err != nil {
			t.Errorf("Save game file error. File :%v, Error: %v", fileName, err.Error())
		} else {
			file, err = os.Open(fileName)
			loadGame, err := Load(file)
			if err != nil {
				t.Errorf("Load game file error. File :%v, Error: %v", fileName, err.Error())
			} else {
				compGames(game, loadGame, t)
			}
		}
	} else {
		t.Errorf("Create file error. File :%v, Error: %v", fileName, err.Error())
	}
}
func TestDecodeGame(t *testing.T) {
	GobRegistor()
	game := New(1, [2]int{1, 2})
	game.Start(0)
	b := new(bytes.Buffer)

	e := gob.NewEncoder(b)

	// Encoding the map
	err := e.Encode(game)
	if err != nil {
		t.Errorf("Error encoding: %v", err)
	}

	var loadGame Game
	d := gob.NewDecoder(b)

	// Decoding the serialized data
	err = d.Decode(&loadGame)
	if err != nil {
		t.Errorf("Error decoding: %v", err)
	} else {
		compGames(game, &loadGame, t)
	}
}
func compGames(save *Game, load *Game, t *testing.T) {
	if !save.Pos.Turn.Equal(&load.Pos.Turn) {
		t.Logf("turn:\n%v\n load turn:\n%v", save.Pos.Turn, load.Pos.Turn)
		t.Error("Save and load turn not equal")
	} else if !save.Pos.Equal(&load.Pos) {
		t.Logf("Pos.Hands:\n%v\n, load pos.Hands:\n%v", save.Pos.Hands, load.Pos.Hands)
		t.Error("Save and load pos not equal")
	} else if !save.Equal(load) {
		t.Logf("Before:\n%v\n", *save)
		t.Logf("After:\n%v", *load)
		t.Error("Save and load game not equal")
	}
}
func TestGame(t *testing.T) {
	fileNameInit := "test/initgame.gob"
	fileNameEnd := "test/endgame.gob"

	/*tmp := New(1, [2]int{1, 2})
	tmp.Start(1)
	GobRegistor()
	saveGame(fileNameInit, tmp)
	*/

	fileInit, err := os.Open(fileNameInit)
	if err == nil {
		defer fileInit.Close()
		GobRegistor()
		game, err := Load(fileInit)
		fileInit.Close()
		if err != nil {
			t.Errorf("Load game file error. File :%v, Error: %v", fileNameInit, err.Error())
		} else {
			var move Move
			move = MoveCardFlag(2)
			game.MoveHand(59, game.Pos.GetMoveix(59, move)) //59,3
			move = MoveDeck(DECK_TROOP)
			game.Move(game.Pos.GetMoveix(0, move)) //0,1

			moveHandDeck(game, 40, 2, DECK_TROOP)  //0
			moveHandDeck(game, 2, 8, DECK_TROOP)   //1
			moveHandDeck(game, 3, 8, DECK_TROOP)   //0
			moveHandDeck(game, 42, 8, DECK_TROOP)  //1
			moveHandDeck(game, 15, 7, DECK_TROOP)  //0
			moveHandDeck(game, 60, 2, DECK_TROOP)  //1
			moveHandDeck(game, 39, 2, DECK_TROOP)  //0
			moveHandDeck(game, 16, 7, DECK_TROOP)  //1
			moveHandDeck(game, 20, 5, DECK_TROOP)  //0
			moveHandDeck(game, 46, 7, DECK_TROOP)  //1
			moveHandDeck(game, 13, 7, DECK_TROOP)  //0
			moveHandDeck(game, 33, 5, DECK_TAC)    //1
			moveHandDeck(game, 14, 7, DECK_TROOP)  //0 Formation
			moveHandDeck(game, 32, 5, DECK_TROOP)  //1
			move = MoveClaim([]int{7})             //0
			game.Move(game.Pos.GetMoveix(0, move)) //0
			moveHandDeck(game, 53, 8, DECK_TROOP)  //0
			moveHandDeck(game, 9, 6, DECK_TROOP)   //1
			moveHandDeck(game, 29, 6, DECK_TROOP)  //0
			moveHandDeck(game, 25, 0, DECK_TROOP)  //1
			moveHandDeck(game, 30, 6, DECK_TAC)    //0
			moveHandDeck(game, 27, 0, DECK_TAC)    //1
			moveHandDeck(game, 44, 4, DECK_TROOP)  //0
			moveHandDeck(game, 5, 1, DECK_TROOP)   //1
			moveHandDeck(game, 18, 5, DECK_TROOP)  //0
			moveHandDeck(game, 22, 8, DECK_TROOP)  //1
			moveHandDeck(game, 43, 8, DECK_TROOP)  //0
			playerClaimFlag(game, 8)               // failed claim
			moveHandDeck(game, 10, 6, DECK_TROOP)  //1
			playerClaimFlag(game, 8)               //0
			moveHandDeck(game, 24, 4, DECK_TROOP)  //0
			moveHandDeck(game, 70, 6, DECK_TROOP)  //1
			//Deserter 62
			move = MoveDeserter{6, 70} // 0
			game.MoveHand(62, game.Pos.GetMoveix(62, move))
			move = MoveDeck(DECK_TAC)
			game.Move(game.Pos.GetMoveix(0, move)) //0
			moveHandDeck(game, 51, 3, DECK_TAC)    //1
			moveHandDeck(game, 34, 3, DECK_TROOP)  //0
			moveHandDeck(game, 6, 1, DECK_TROOP)   //1
			moveHandDeck(game, 4, 4, DECK_TROOP)   //0
			moveHandDeck(game, 7, 1, DECK_TROOP)   //1
			playerClaimFlag(game, -1)              // No claim 0
			moveHandDeck(game, 28, 6, DECK_TROOP)  //0
			playerClaimFlag(game, -1)              // No claim 1
			//Fog
			moveHandDeck(game, 66, 6, DECK_TROOP)  //1
			playerClaimFlag(game, -1)              //0
			moveHandDeck(game, 36, 3, DECK_TROOP)  //0
			playerClaimFlag(game, -1)              //1
			moveHandDeck(game, 49, 6, DECK_TROOP)  //1
			playerClaimFlag(game, -1)              //0
			moveHandDeck(game, 1, 0, DECK_TAC)     //0
			move = MoveClaim([]int{1, 6})          // claim two get one//1
			game.Move(game.Pos.GetMoveix(0, move)) //1
			moveHandDeck(game, 21, 3, DECK_TROOP)  //1
			playerClaimFlag(game, -1)              //0
			moveHandDeck(game, 57, 1, DECK_TAC)    //0
			playerClaimFlag(game, -1)              //1
			moveHandDeck(game, 58, 2, DECK_TROOP)  //1
			playerClaimFlag(game, -1)              //0
			//Mud
			moveHandDeck(game, 65, 1, DECK_TROOP) //0
			playerClaimFlag(game, 2)              //1
			moveHandDeck(game, 48, 1, DECK_TROOP) //1
			playerClaimFlag(game, -1)             //0
			//scout
			move = MoveDeck(DECK_TROOP) //0
			game.MoveHand(64, game.Pos.GetMoveix(64, move))
			move = MoveDeck(DECK_TROOP)
			game.Move(game.Pos.GetMoveix(-1, move))
			move = MoveDeck(DECK_TROOP)
			game.Move(game.Pos.GetMoveix(-1, move))
			move = MoveScoutReturn{[]int{67}, []int{11}}
			game.Move(game.Pos.GetMoveix(-1, move))
			playerClaimFlag(game, 1)               // fail 1
			moveHandDeck(game, 41, 3, DECK_TAC)    //1
			playerClaimFlag(game, -1)              //0
			moveHandDeck(game, 19, 5, DECK_TROOP)  //0
			playerClaimFlag(game, 1)               // 1
			moveHandDeck(game, 26, 0, DECK_TAC)    //1
			move = MoveClaim([]int{4, 5})          // claim two get one//0
			game.Move(game.Pos.GetMoveix(0, move)) //0
			moveHandDeck(game, 11, 0, DECK_TROOP)  //0
			playerClaimFlag(game, 0)               // 1
			moveHandDeck(game, 55, 4, DECK_TROOP)  //1
			playerClaimFlag(game, -1)              //0
			moveHandDeck(game, 47, 1, DECK_TROOP)  //0
			playerClaimFlag(game, -1)              // 1
			moveHandDeck(game, 54, 4, DECK_TROOP)  //1
			playerClaimFlag(game, -1)              //0
			moveHandDeck(game, 17, 1, DECK_TROOP)  //0
			playerClaimFlag(game, -1)              // 1
			moveHandDeck(game, 56, 4, DECK_TAC)    //1
			playerClaimFlag(game, -1)              //0
			moveHandDeck(game, 38, 3, DECK_TAC)    //0
			move = MoveClaim([]int{3, 4})          //1 claim two get  two and win
			game.Move(game.Pos.GetMoveix(0, move)) //1

			_ = fmt.Sprintf("Move pos: %v\n", game.Pos) // Do not want to delete fmt

			/*fmt.Printf("Move pos: %v\n", game.Pos)
			fmt.Printf("Flag 6: %v\n", game.Pos.Flags[7].Players[0])
			fmt.Printf("Flag 9: %v\n", game.Pos.Flags[8].Players[1])
			fmt.Printf("Flag 7: %v\n", game.Pos.Flags[6].Players[1])
			fmt.Printf("Flag 2: %v\n", game.Pos.Flags[1].Players[1])
			fmt.Printf("Moves player 0: %v\n", game.Pos.MovesHand[62])
			fmt.Printf("MovesHands: %v\n", game.Pos.MovesHand)
			fmt.Printf("Moves: %v\n", game.Pos.Moves)
			fmt.Printf("Tac Deck: %v\n", game.Pos.DeckTac)
			*/

			//saveGame(fileNameEnd, game)

			fileEnd, err := os.Open(fileNameEnd)
			defer fileEnd.Close()
			if err == nil {
				gameEnd, err := Load(fileEnd)
				if err != nil {
					t.Errorf("Error loading file :%v. Error:%v", fileNameEnd, err.Error())
				} else {
					compGames(game, gameEnd, t)
				}
			} else {
				t.Errorf("Error open file :%v. Error:%v", fileNameEnd, err.Error())
			}

		}
	} else {
		t.Errorf("Open file error. File :%v, Error: %v", fileNameInit, err.Error())
	}
}
func playerClaimFlag(game *Game, flag int) {
	var move Move
	if flag == -1 {
		move = MoveClaim([]int{}) //make([]int,0)
	} else {
		move = MoveClaim([]int{flag})
	}
	game.Move(game.Pos.GetMoveix(0, move))

}
func moveHandDeck(game *Game, card int, flag int, deck int) {
	var move Move
	move = MoveCardFlag(flag)
	ix := game.Pos.GetMoveix(card, move)
	game.MoveHand(card, ix)
	move = MoveDeck(deck)
	game.Move(game.Pos.GetMoveix(0, move))

}
func saveGame(fileName string, game *Game) (err error) {
	file, err := os.Create(fileName)
	if err == nil {
		GobRegistor()
		err = Save(game, file)
		file.Close()
	}
	return err
}
func TestGameTraitor(t *testing.T) {
	fileNameInit := "test/initgametraitor.gob"
	fileNameEnd := "test/endgametraitor.gob"

	fileInit, err := os.Open(fileNameInit)
	if err == nil {
		defer fileInit.Close()
		GobRegistor()
		game, err := Load(fileInit)
		fileInit.Close()
		if err != nil {
			t.Errorf("Load game file error. File :%v, Error: %v", fileNameInit, err.Error())
		} else {
			var move Move
			moveHandDeck(game, 50, 3, DECK_TROOP) //1
			moveHandDeck(game, 48, 3, DECK_TROOP) //0
			moveHandDeck(game, 58, 5, DECK_TROOP) //1
			moveHandDeck(game, 47, 3, DECK_TROOP) //0
			moveHandDeck(game, 57, 5, DECK_TAC)   //1
			moveHandDeck(game, 46, 3, DECK_TAC)   //0
			//Redeploy
			move = MoveRedeploy{3, 50, 4}                   //1
			game.MoveHand(63, game.Pos.GetMoveix(63, move)) //1
			move = MoveDeck(DECK_TROOP)                     //1
			game.Move(game.Pos.GetMoveix(0, move))          //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 19, 2, DECK_TROOP)           //0
			moveHandDeck(game, 10, 4, DECK_TROOP)           //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 38, 4, DECK_TROOP)           //0
			moveHandDeck(game, 56, 5, DECK_TROOP)           //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 11, 0, DECK_TROOP)           //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 41, 0, DECK_TROOP)           //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 36, 4, DECK_TROOP)           //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 21, 0, DECK_TROOP)           //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 12, 0, DECK_TROOP)           //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 27, 6, DECK_TROOP)           //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 8, 6, DECK_TROOP)            //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 1, 0, DECK_TROOP)            //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 6, 6, DECK_TROOP)            //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 17, 6, DECK_TROOP)           //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 32, 8, DECK_TROOP)           //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 37, 6, DECK_TROOP)           //1
			playerClaimFlag(game, -1)                       //0
			//traitor
			move = MoveTraitor{6, 37, 4}                    //0
			game.MoveHand(61, game.Pos.GetMoveix(61, move)) //0
			move = MoveDeck(DECK_TROOP)                     //0
			game.Move(game.Pos.GetMoveix(0, move))          //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 13, 8, DECK_TROOP)           //1
			playerClaimFlag(game, 4)                        //0
			moveHandDeck(game, 33, 8, DECK_TAC)             //0
			playerClaimFlag(game, 0)                        //1
			moveHandDeck(game, 15, 7, DECK_TROOP)           //1
			playerClaimFlag(game, 3)                        //0 fail
			moveHandDeck(game, 49, 2, DECK_TROOP)           //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 35, 7, DECK_TROOP)           //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 59, 2, DECK_TROOP)           //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 45, 7, DECK_TAC)             //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 23, 7, DECK_TROOP)           //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 18, 8, DECK_TROOP)           //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 29, 5, DECK_TROOP)           //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 34, 3, DECK_TAC)             //1
			playerClaimFlag(game, 3)                        //0
			moveHandDeck(game, 24, 7, DECK_TROOP)           //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 5, 2, DECK_TROOP)            //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 30, 5, DECK_TROOP)           //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 60, 1, DECK_TROOP)           //1
			playerClaimFlag(game, -1)                       //0
			moveHandDeck(game, 28, 5, DECK_TROOP)           //0
			playerClaimFlag(game, -1)                       //1
			moveHandDeck(game, 40, 1, DECK_TROOP)           //1
			playerClaimFlag(game, 5)                        //0 win with 3
			//fmt.Printf("Move pos: %v\n", game.Pos)
			//fmt.Printf("Flag 6: %v\n", game.Pos.Flags[7].Players[0])
			//fmt.Printf("Flag 9: %v\n", game.Pos.Flags[8].Play:ers[1])
			//fmt.Printf("Flag 7: %v\n", game.Pos.Flags[6].Players[1])
			//fmt.Printf("Flag 2: %v\n", game.Pos.Flags[1].Players[1])
			//fmt.Printf("Moves player 0: %v\n", game.Pos.MovesHand[62])
			//fmt.Printf("Turn player:%v\n", game.Pos.Player)
			//fmt.Printf("Hand 0:%v,Hand 1:%v\n", game.Pos.Hands[0], game.Pos.Hands[1])
			//fmt.Printf("MovesHands: %v\n", game.Pos.MovesHand)
			//fmt.Printf("Moves: %v\n", game.Pos.Moves)
			//fmt.Printf("Tac Deck: %v\n", game.Pos.DeckTac)
			//fmt.Printf("Troop Deck: %v\n", game.Pos.DeckTroop)

			//saveGame(fileNameEnd, game)

			fileEnd, err := os.Open(fileNameEnd)
			defer fileEnd.Close()
			if err == nil {
				gameEnd, err := Load(fileEnd)
				if err != nil {
					t.Errorf("Error loading file :%v. Error:%v", fileNameEnd, err.Error())
				} else {
					compGames(game, gameEnd, t)
				}
			} else {
				t.Errorf("Error open file :%v. Error:%v", fileNameEnd, err.Error())
			}

		}
	}
}
