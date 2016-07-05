// A server for battleline
package tables

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	bat "rezder.com/game/card/battleline"
	"rezder.com/game/card/battleline/flag"
	pub "rezder.com/game/card/battleline/server/publist"
	slice "rezder.com/slice/int"
	"time"
)

//Start a table with a game.
//If resumeGame is nil a new game is started.
func table(ids [2]int, playerChs [2]chan<- *pub.MoveView, watchChCl *pub.WatchChCl, resumeGame *bat.Game,
	finishCh chan *FinishTableData, save bool, savedir string, errCh chan<- error) {

	var moveChs [2]chan [2]int
	moveChs[0] = make(chan [2]int)
	moveChs[1] = make(chan [2]int)
	benchCh := make(chan *MoveBenchView, 1)
	go bench(watchChCl, benchCh)
	game, moveInit1, moveInit2, moveInitBench := initMove(resumeGame, ids, moveChs)
	playerChs[0] <- moveInit1
	playerChs[1] <- moveInit2
	benchCh <- moveInitBench
	var moveix [2]int
	var move bat.Move
	var open bool
	var mover int

	var deltCardix int
	var claimFailMap map[string][]int //may contain zero index for 4 card
	var isScout bool
	var scout bat.MoveScoutReturn
	var isClaim bool
	var claim bat.MoveClaim
	var claimView *MoveClaimView
	var redeploy bat.MoveRedeploy
	var isRedeploy bool
	var redeployDishixs []int
	var isSaveMove = false
	for {
		isScout = false
		isClaim = false
		isRedeploy = false
		isSaveMove = false
		deltCardix = 0
		claimFailMap = nil
		mover = game.Pos.Player
		fmt.Printf("Waiting for mover ix: %v id: %v\n", mover, ids[mover])
		moveix, open = <-moveChs[mover]
		fmt.Printf("Recived move ix:%v from mover ix: %v id:%v\n", moveix, mover, ids[mover])
		if !open {
			isSaveMove = true
			move = *NewMoveSave()
		} else if moveix[1] == pub.SM_Quit {
			game.Quit(game.Pos.Player)
			move = *NewMoveQuit()
		} else if moveix[1] == pub.SM_Pass {
			game.Pass()
			move = *NewMovePass()
		} else if moveix[0] > 0 {
			move = game.Pos.MovesHand[moveix[0]][moveix[1]]
			deltCardix, redeployDishixs = game.MoveHand(moveix[0], moveix[1])
			redeploy, isRedeploy = move.(bat.MoveRedeploy)
			if isRedeploy {
				move = *NewMoveRedeployView(&redeploy, redeployDishixs)
			}
		} else {
			move = game.Pos.Moves[moveix[1]]
			scout, isScout = move.(bat.MoveScoutReturn)
			if isScout {
				move = *NewMoveScoutReturnView(scout)
			} else {
				claim, isClaim = move.(bat.MoveClaim)
				if isClaim {
					claimView = NewMoveClaimView(claim)
				}
			}
			deltCardix, claimFailMap = game.Move(moveix[1])
		}
		if isClaim {
			move = updateClaim(game.Pos, claimFailMap, claimView)
		}
		move1, move2, moveBench := creaMove(mover, move, moveix[0], deltCardix, game.Pos, ids)
		fmt.Printf("Sending move to playerid: %v\n%v\n", ids[0], move1)
		playerChs[0] <- move1
		fmt.Printf("Sending move to playerid: %v\n%v\n", ids[1], move2)
		playerChs[1] <- move2
		benchCh <- moveBench
		if game.Pos.State == bat.TURN_FINISH || game.Pos.State == bat.TURN_QUIT || isSaveMove {
			if save {
				hour, min, _ := time.Now().Clock()
				fileName := fmt.Sprintf("game%vvs%v%v%v.gob", ids[0], ids[1], hour, min)
				fileNamePath := filepath.Join(savedir, fileName)
				file, err := os.Create(fileNamePath)
				//defer file.Close()//close even if panic. Double close produce a error
				//but we are not listening it is possible to add the close error to err of a returning function
				//but id do not think i care.
				if err == nil {
					err = bat.Save(game, file, true)
					file.Close()
					if err != nil {
						errCh <- err
					}
				} else {
					errCh <- err
				}
			}
			break
		}
	}

	close(playerChs[0])
	close(playerChs[1])
	close(benchCh)
	finData := new(FinishTableData)
	finData.ids = ids
	if isSaveMove {
		finData.game = game
	}
	finishCh <- finData // may be buffered tables will close bench before closing tables
}

//updateClaim update the MoveClaimView with the succesfully claim flags and the win
//indicator. The updated view is casted and returned as the move.
func updateClaim(pos *bat.GamePos, claimFailMap map[string][]int, claimView *MoveClaimView) (move bat.Move) {

	for _, v := range claimView.Claim {
		if pos.Flags[v].Claimed() {
			claimView.Claimed = append(claimView.Claimed, v)
		}
	}
	if len(claimFailMap) != 0 {
		claimView.FailMap = claimFailMap
	}
	if pos.State == bat.TURN_FINISH {
		claimView.Win = true
	}
	move = *claimView //TODO this works with ref and without we need to find out why and when we
	// should use ref currently interface moves value is not ref but could be??
	return move
}

//creaMove Create a player move.
func creaMove(mover int, move bat.Move, moveCardix int, deltCardix int, pos *bat.GamePos, ids [2]int) (move1 *pub.MoveView, move2 *pub.MoveView, moveBench *MoveBenchView) {
	move1 = new(pub.MoveView)
	move1.Mover = mover == 0
	move1.Move = move
	move1.MoveCardix = moveCardix
	move1.Turn = pub.NewTurn(&pos.Turn, 0)

	move2 = new(pub.MoveView)
	move2.Mover = mover == 1
	move2.Move = move
	move2.MoveCardix = moveCardix
	move2.Turn = pub.NewTurn(&pos.Turn, 1)

	if deltCardix != 0 {
		if mover == 0 {
			move1.DeltCardix = deltCardix
		} else {
			move2.DeltCardix = deltCardix
		}
	}
	moveBench = new(MoveBenchView)
	moveBench.Mover = mover
	moveBench.Move = move
	moveBench.NextMover = pos.Player
	moveBench.MoveCardix = moveCardix
	initMove := NewMoveBenchPos(NewBenchPos(pos, ids))
	moveBench.MoveInit = initMove

	return move1, move2, moveBench
}

//initMove create the initial moves.
func initMove(resumeGame *bat.Game, ids [2]int, moveChans [2]chan [2]int) (game *bat.Game, move1 *pub.MoveView,
	move2 *pub.MoveView, moveBench *MoveBenchView) {
	move1 = new(pub.MoveView)
	move2 = new(pub.MoveView)
	moveBench = new(MoveBenchView)
	moveBench.Mover = -1

	if resumeGame != nil {
		game = resumeGame
		if resumeGame.PlayerIds[0] != ids[0] || resumeGame.PlayerIds[1] != ids[1] {
			panic("Old game and player id out of synch")
		}
		move1.Turn = pub.NewTurn(&game.Pos.Turn, 0)
		move1.MoveCh = moveChans[0]
		move1.Move = NewMoveInitPos(NewPlayPos(game.Pos, 0))

		move2.Turn = pub.NewTurn(&game.Pos.Turn, 1)
		move2.MoveCh = moveChans[1]
		move2.Move = NewMoveInitPos(NewPlayPos(game.Pos, 1))

		moveBench.NextMover = game.Pos.Player
		moveBench.MoveInit = NewMoveBenchPos(NewBenchPos(game.Pos, ids))
	} else {
		game = bat.New(ids)
		rand.Seed(time.Now().UnixNano())
		r := rand.Perm(2)
		game.Start(r[0])

		move1.Turn = pub.NewTurn(&game.Pos.Turn, 0)
		move1.MoveCh = moveChans[0]
		h := make([]int, len(game.Pos.Hands[0].Troops))
		copy(h, game.Pos.Hands[0].Troops)
		move1.Move = NewMoveInit(h)

		move2.Turn = pub.NewTurn(&game.Pos.Turn, 1)
		move2.MoveCh = moveChans[1]
		h = make([]int, len(game.Pos.Hands[1].Troops))
		copy(h, game.Pos.Hands[1].Troops)
		move2.Move = NewMoveInit(h)

		moveBench.NextMover = game.Pos.Player
		moveBench.Move = NewMoveBenchInit(ids)
	}
	return game, move1, move2, moveBench
}

//PlayPos a player position for starting a old game.
type PlayPos struct {
	Flags         [bat.FLAGS]*Flag
	OppDishTroops []int
	OppDishTacs   []int
	DishTroops    []int
	DishTacs      []int
	OppHand       []bool //true equal troop false equal tactic.
	Hand          []int
	DeckTacs      int
	DeckTroops    int
}

//opp return the oppont.
func opp(playerix int) int {
	if playerix == 0 {
		return 1
	} else {
		return 0
	}
}
func NewPlayPos(pos *bat.GamePos, playerix int) (posView *PlayPos) {
	opp := opp(playerix)
	posView = new(PlayPos)
	for i, v := range pos.Flags {
		posView.Flags[i] = NewFlag(v, playerix)
	}

	posView.OppDishTroops, posView.OppDishTacs = copyDish(pos, opp)
	posView.DishTroops, posView.DishTacs = copyDish(pos, playerix)

	troops := len(pos.Hands[opp].Troops)
	posView.OppHand = make([]bool, troops+len(pos.Hands[opp].Tacs))
	for i, _ := range pos.Hands[opp].Troops {
		posView.OppHand[i] = true
	}
	for i, _ := range pos.Hands[opp].Tacs {
		posView.OppHand[i+troops] = false
	}
	troops = len(pos.Hands[playerix].Troops)
	posView.Hand = make([]int, troops+len(pos.Hands[playerix].Tacs))
	for i, v := range pos.Hands[playerix].Troops {
		posView.Hand[i] = v
	}
	for i, v := range pos.Hands[playerix].Tacs {
		posView.Hand[i+troops] = v
	}
	posView.DeckTroops = len(pos.DeckTroop.Remaining())
	posView.DeckTacs = len(pos.DeckTac.Remaining())

	return posView
}

//copyDish extract the dished card data.
func copyDish(pos *bat.GamePos, ix int) (troops []int, tacs []int) {
	troops = make([]int, len(pos.Dishs[ix].Troops))
	copy(troops, pos.Dishs[ix].Troops)
	tacs = make([]int, len(pos.Dishs[ix].Tacs))
	copy(tacs, pos.Dishs[ix].Tacs)
	return troops, tacs
}

func (pp *PlayPos) Equal(other *PlayPos) (equal bool) {
	if other == nil && pp == nil {
		equal = true
	} else if other != nil && pp != nil {
		if pp == other {
			equal = true
		} else if pp.DeckTacs == other.DeckTacs && pp.DeckTroops == other.DeckTroops &&
			slice.Equal(pp.OppDishTroops, other.OppDishTroops) && slice.Equal(pp.OppDishTacs, other.OppDishTacs) &&
			slice.Equal(pp.DishTroops, other.DishTroops) &&
			slice.Equal(pp.DishTacs, other.DishTacs) && slice.Equal(pp.Hand, other.Hand) {
			equal = true
			if len(pp.OppHand) != len(other.OppHand) {
				equal = false
			} else {
				if len(pp.OppHand) > 0 {
					for i, v := range pp.OppHand {
						if v != other.OppHand[i] {
							equal = false
							break
						}
					}

				}
			}
			if equal {
				for i, v := range other.Flags {
					if !v.Equal(pp.Flags[i]) {
						equal = false
						break
					}
				}
			}
		}
	}
	return equal
}

// Flag a battleline flag.
type Flag struct {
	OppFlag    bool
	OppTroops  []int
	OppEnvs    []int
	NeuFlag    bool
	PlayEnvs   []int
	PlayTroops []int
	PlayFlag   bool
}

func NewFlag(flag *flag.Flag, playerix int) (view *Flag) {
	opp := opp(playerix)
	view = new(Flag)
	view.OppFlag = flag.Players[opp].Won
	view.OppTroops, view.OppEnvs = flagCopyCards(flag, opp)
	view.NeuFlag = !flag.Players[0].Won && !flag.Players[1].Won
	view.PlayTroops, view.PlayEnvs = flagCopyCards(flag, playerix)
	view.PlayFlag = flag.Players[playerix].Won
	return view
}

//flagCopyCards extract cards of a player.
func flagCopyCards(flag *flag.Flag, playerix int) (troops []int, envs []int) {
	troops = make([]int, 0, len(flag.Players[playerix].Troops))
	for _, v := range flag.Players[playerix].Troops {
		if v != 0 {
			troops = append(troops, v)
		}
	}
	envs = make([]int, 0, len(flag.Players[playerix].Env))
	for _, v := range flag.Players[playerix].Env {
		if v != 0 {
			envs = append(envs, v)
		}
	}
	return troops, envs
}
func (flag *Flag) Equal(other *Flag) (equal bool) {
	if other == nil && flag == nil {
		equal = true
	} else if other != nil && flag != nil {
		if flag == other {
			equal = true
		} else if flag.PlayFlag == other.PlayFlag && flag.NeuFlag == other.NeuFlag &&
			flag.OppFlag == other.OppFlag && slice.Equal(flag.OppTroops, other.OppTroops) &&
			slice.Equal(flag.OppEnvs, other.OppEnvs) && slice.Equal(flag.PlayEnvs, other.PlayEnvs) &&
			slice.Equal(flag.PlayTroops, other.PlayTroops) {
			equal = true
		}
	}
	return equal
}

// MoveReployView the redeploy move view.
type MoveRedeployView struct {
	Move            *bat.MoveRedeploy
	RedeployDishixs []int
	JsonType        string
}

func NewMoveRedeployView(move *bat.MoveRedeploy, dishixs []int) (m *MoveRedeployView) {
	m = new(MoveRedeployView)
	m.Move = move
	m.RedeployDishixs = dishixs
	m.JsonType = "MoveRedeployView"
	return m
}

func (m MoveRedeployView) MoveEqual(other bat.Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveRedeployView)
		if ok && m.Move.MoveEqual(om.Move) && slice.Equal(m.RedeployDishixs, om.RedeployDishixs) {
			equal = true
		}
	}
	return equal
}
func (m MoveRedeployView) Copy() (c bat.Move) {
	c = m
	return c
}

// MoveScoutReturnView the scout return move view.
// the scout return move as view by the opponent and the public.
type MoveScoutReturnView struct {
	Tac      int
	Troop    int
	JsonType string
}

func NewMoveScoutReturnView(scout bat.MoveScoutReturn) (m *MoveScoutReturnView) {
	m = new(MoveScoutReturnView)
	m.Troop = len(scout.Troop)
	m.Tac = len(scout.Tac)
	m.JsonType = "MoveScoutReturnView"
	return m
}

func (m MoveScoutReturnView) MoveEqual(other bat.Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveScoutReturnView)
		if ok && om.Tac == m.Tac && om.Troop == m.Troop {
			equal = true
		}
	}
	return equal
}
func (m MoveScoutReturnView) Copy() (c bat.Move) {
	c = m
	return c
}

// MoveClaimView the claim view contain the opriginal claim move and the result.
// The claimed flags, if the game was won and evt. information from failed claims.
type MoveClaimView struct {
	Claim    []int //The players claimed this flags
	Claimed  []int //The players claimed flag that was not rejected
	Win      bool
	FailMap  map[string][]int
	JsonType string
}

func (m MoveClaimView) Equal(other MoveClaimView) (equal bool) {
	if m.Win == other.Win && slice.Equal(m.Claim, other.Claim) &&
		slice.Equal(m.Claimed, other.Claimed) {
		equal = true
		if len(m.FailMap) == len(other.FailMap) {
			if len(m.FailMap) != 0 {
				for flagix, ex := range m.FailMap {
					otherEx, found := other.FailMap[flagix]
					if found {
						if !slice.Equal(ex, otherEx) {
							equal = false
							break
						}
					} else {
						equal = false
						break
					}
				}
			} else {
				equal = true
			}
		} else {
			equal = false
		}
	}
	return equal
}
func (m MoveClaimView) MoveEqual(other bat.Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveClaimView)
		if ok {
			equal = m.Equal(om)
		}
	}
	return equal
}
func (m MoveClaimView) Copy() (c bat.Move) {
	c = m //no deep copy
	return c
}

func NewMoveClaimView(claim bat.MoveClaim) (m *MoveClaimView) {
	m = new(MoveClaimView)
	m.Claim = make([]int, len(claim.Flags))
	m.Claimed = make([]int, 0, len(claim.Flags))
	copy(m.Claim, claim.Flags)
	m.JsonType = "MoveClaimView"
	return m
}

//MoveInit the initial move in a new game.
//Just the 7 cards the player get.
type MoveInit struct {
	Hand     []int //nil for bench
	JsonType string
}

func NewMoveInit(hand []int) *MoveInit {
	res := new(MoveInit)
	res.Hand = hand
	res.JsonType = "MoveInit"
	return res
}

func (m MoveInit) Equal(other MoveInit) (equal bool) {
	if slice.Equal(other.Hand, m.Hand) {
		equal = true
	}
	return equal
}

func (m MoveInit) MoveEqual(other bat.Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveInit)
		if ok {
			equal = m.Equal(om)
		}
	}
	return equal
}
func (m MoveInit) Copy() (c bat.Move) {
	c = m //no deep copy
	return c
}

//MoveInitPos is the first move in a resumed game.
type MoveInitPos struct {
	Pos      *PlayPos
	JsonType string
}

func NewMoveInitPos(pos *PlayPos) *MoveInitPos {
	res := new(MoveInitPos)
	res.Pos = pos
	res.JsonType = "MoveInitPos"
	return res
}
func (m MoveInitPos) MoveEqual(other bat.Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveInitPos)
		if ok && m.Pos.Equal(om.Pos) {
			equal = true
		}
	}
	return equal
}
func (m MoveInitPos) Copy() (c bat.Move) {
	c = m //no deep copy
	return c
}

// MovePass is the pass move.
type MovePass struct {
	JsonType string
}

func NewMovePass() *MovePass {
	res := new(MovePass)
	res.JsonType = "MovePass"
	return res
}

func (m MovePass) MoveEqual(other bat.Move) (equal bool) {
	if other != nil {
		_, ok := other.(MovePass)
		if ok {
			equal = true
		}
	}
	return equal
}
func (m MovePass) Copy() (c bat.Move) {
	c = m
	return c
}

//MoveQuit the quit move.
type MoveQuit struct {
	JsonType string
}

func NewMoveQuit() *MoveQuit {
	res := new(MoveQuit)
	res.JsonType = "MoveQuit"
	return res
}
func (m MoveQuit) MoveEqual(other bat.Move) (equal bool) {
	if other != nil {
		_, ok := other.(MoveQuit)
		if ok {
			equal = true
		}
	}
	return equal
}
func (m MoveQuit) Copy() (c bat.Move) {
	c = m
	return c
}

//MoveSave the save move.
type MoveSave struct {
	JsonType string
}

func NewMoveSave() *MoveSave {
	res := new(MoveSave)
	res.JsonType = "MoveSave"
	return res
}
func (m MoveSave) MoveEqual(other bat.Move) (equal bool) {
	if other != nil {
		_, ok := other.(MoveSave)
		if ok {
			equal = true
		}
	}
	return equal
}
func (m MoveSave) Copy() (c bat.Move) {
	c = m
	return c
}
