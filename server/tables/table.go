// A server for battleline
package tables

import (
	"math/rand"
	bat "rezder.com/game/card/battleline"
	"rezder.com/game/card/battleline/flag"
	pub "rezder.com/game/card/battleline/server/publist"
	slice "rezder.com/slice/int"
	"time"
)

func table(ids [2]int, playerChans [2]chan<- *pub.MoveView, watch *pub.WatchChan, resumeGame *bat.Game,
	finish chan<- [2]int) {

	var moveChans [2]chan [2]int
	moveChans[0] = make(chan [2]int)
	moveChans[1] = make(chan [2]int)
	benchChan := make(chan *MoveBenchView, 1)
	go bench(watch, benchChan)
	game, moveInit1, moveInit2, moveInitBench := initMove(resumeGame, ids, moveChans)
	playerChans[0] <- moveInit1
	playerChans[1] <- moveInit2
	benchChan <- moveInitBench

	var moveix [2]int
	var move bat.Move
	var open bool
	var mover int

	var cardix int
	var isScout bool
	var scout bat.MoveScoutReturn
	var isClaim bool
	var claim bat.MoveClaim
	var claimView *MoveClaimView
	for {
		isScout = false
		isClaim = false
		cardix = 0
		mover = game.Pos.Player
		moveix, open = <-moveChans[mover]
		if !open {
			game.Quit(game.Pos.Player)
			move = MoveQuit{}
		} else if moveix[0] == -1 && moveix[1] == -1 {
			game.Pass()
			move = MovePass{}
		} else if moveix[0] != -1 {
			move = game.Pos.MovesHand[moveix[0]][moveix[1]]
			cardix = game.MoveHand(moveix[0], moveix[1])
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
			cardix = game.Move(moveix[1])
		}
		if isClaim {
			move = updateClaim(&game.Pos, claimView)
		}
		move1, move2, moveBench := creaMove(mover, move, cardix, &game.Pos)
		playerChans[0] <- move1
		playerChans[1] <- move2
		benchChan <- moveBench
		if game.Pos.State == bat.TURN_FINISH || game.Pos.State == bat.TURN_QUIT {
			break
		}
	}

	close(playerChans[0])
	close(playerChans[1])
	finish <- ids // may be buffered tables will close bench before closing tables
	//close(benchChan) finish close the watch that close the bench

}
func updateClaim(pos *bat.GamePos, claimView *MoveClaimView) (move bat.Move) {
	for _, v := range claimView.Claim {
		if pos.Flags[v].Claimed() {
			claimView.Claimed = append(claimView.Claimed, v)
		}
	}
	if len(pos.Info) != 0 {
		claimView.Info = pos.Info
	}
	if pos.State == bat.TURN_FINISH {
		claimView.Win = true
	}
	move = *claimView //TODO this works with ref and without we need to find out why and when we
	// should use ref currently interface moves value is not ref but could be??
	return move
}
func creaMove(mover int, move bat.Move, cardix int, pos *bat.GamePos) (move1 *pub.MoveView, move2 *pub.MoveView, moveBench *MoveBenchView) {
	move1 = new(pub.MoveView)
	move1.Mover = mover == 0
	move1.Move = move
	move1.Turn = pub.NewTurn(&pos.Turn, 0)

	move2 = new(pub.MoveView)
	move2.Mover = mover == 1
	move2.Move = move
	move2.Turn = pub.NewTurn(&pos.Turn, 1)

	if cardix != 0 {
		if mover == 0 {
			move1.Card = cardix
		} else {
			move2.Card = cardix
		}
	}
	moveBench = new(MoveBenchView)
	moveBench.Mover = mover
	moveBench.Move = move
	moveBench.NextMover = pos.Player
	initMove := *new(MoveBenchPos)
	initMove.Pos = NewBenchPos(pos)
	moveBench.MoveInit = initMove

	return move1, move2, moveBench
}
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
		moveInitPos := *new(MoveInitPos)
		move1.MoveChan = moveChans[0]
		moveInitPos.Pos = NewPlayPos(&game.Pos, 0)
		move1.Move = moveInitPos

		move2.Turn = pub.NewTurn(&game.Pos.Turn, 1)
		moveInitPos = *new(MoveInitPos)
		move2.MoveChan = moveChans[1]
		moveInitPos.Pos = NewPlayPos(&game.Pos, 1)
		move2.Move = moveInitPos

		moveBench.NextMover = game.Pos.Player
		moveBenchPos := *new(MoveBenchPos)
		moveBenchPos.Pos = NewBenchPos(&game.Pos)
		moveBench.MoveInit = moveBenchPos
	} else {
		game = bat.New(ids)
		rand.Seed(time.Now().UnixNano())
		r := rand.Perm(2)
		game.Start(r[0])

		move1.Turn = pub.NewTurn(&game.Pos.Turn, 0)
		init := new(MoveInit)
		move1.MoveChan = moveChans[0]
		h := make([]int, len(game.Pos.Hands[0].Troops))
		copy(h, game.Pos.Hands[0].Troops)
		init.Hand = h
		move1.Move = init

		move2.Turn = pub.NewTurn(&game.Pos.Turn, 1)
		init = new(MoveInit)
		move2.MoveChan = moveChans[1]
		h = make([]int, len(game.Pos.Hands[1].Troops))
		copy(h, game.Pos.Hands[1].Troops)
		init.Hand = h
		move2.Move = init

		moveBench.NextMover = game.Pos.Player
		init = new(MoveInit)
		moveBench.Move = init
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
	view.OppFlag = flag.Players[opp].Won
	view.OppTroops, view.OppEnvs = flagCopyCards(flag, opp)
	view.NeuFlag = !flag.Players[0].Won && !flag.Players[1].Won
	view.PlayTroops, view.PlayEnvs = flagCopyCards(flag, playerix)
	view.PlayFlag = flag.Players[playerix].Won
	return view
}
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

type MoveScoutReturnView struct {
	Tac   int
	Troop int
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
func NewMoveScoutReturnView(scout bat.MoveScoutReturn) (m *MoveScoutReturnView) {
	m = new(MoveScoutReturnView)
	m.Troop = len(scout.Troop)
	m.Tac = len(scout.Tac)
	return m
}

type MoveClaimView struct {
	Claim   []int
	Claimed []int
	Win     bool
	Info    string
}

func (m MoveClaimView) Equal(other MoveClaimView) (equal bool) {
	if m.Info == other.Info && m.Win == other.Win &&
		slice.Equal(m.Claim, other.Claim) && slice.Equal(m.Claimed, other.Claimed) {
		equal = true
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
	m.Claim = make([]int, len(claim))
	copy(m.Claim, claim)
	return m
}

type MoveInit struct {
	Hand []int //nil for bench
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

type MoveInitPos struct {
	Pos *PlayPos
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

type MovePass struct{}

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

type MoveQuit struct{}

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
