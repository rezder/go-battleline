package flag

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob/combi"
	"github.com/rezder/go-battleline/v2/bot/prob/dht"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-error/log"
	math "github.com/rezder/go-math/int"
)

// Analysis a analysis of the flag.
type Analysis struct {
	Flagix           int
	TargetRank       int
	TargetSum        int
	BotMaxRank       int
	BotMaxSum        int
	OppFormationSize int
	BotFormationSize int
	FormationSize    int
	FlagValue        int
	IsTargetMade     bool
	IsLost           bool
	IsWin            bool
	IsNewFlag        bool
	IsPlayable       bool
	IsClaimed        bool
	Claimer          int
	IsFog            bool
	IsLoosingGame    bool
	RankAnas         []*combi.Analysis
	Flag             *game.Flag
	Playix           int
	//TODO maybe add HostRank
}

func (ana *Analysis) String() string {
	if ana == nil {
		return "<nil>"
	}
	combiAna := "<nil>"
	if ana.RankAnas != nil {
		positive := make([]*combi.Analysis, 0, 0)
		for _, c := range ana.RankAnas {
			if c != nil {
				if c.Prop > 0 {
					positive = append(positive, c)
				}
			}
		}
		if len(positive) == 0 {
			combiAna = "[]"
		} else {
			combiAna = "["
			for i, c := range positive {
				if i == 4 {
					break
				}
				if i != 0 {
					combiAna = combiAna + " "
				}
				combiAna = combiAna + c.String()
			}
			combiAna = combiAna + "]"
		}
	}
	txt := fmt.Sprintf("{Flagix:%v TargetRank:%v BotMaxRank:%v FlagValue:%v IsLost:%v LooseGame:%v Combination:%v",
		ana.Flagix+1, ana.TargetRank, ana.BotMaxRank, ana.FlagValue, ana.IsLost, ana.IsLoosingGame, combiAna)
	return txt
}

// NewAnalysis create a flag analysis.
func NewAnalysis(
	playix int,
	flag *game.Flag,
	deckHand *dht.Cache,
	flagix int) (fa *Analysis) {
	fa = new(Analysis)
	oppix := opp(playix)
	fa.Playix = playix
	fa.Flagix = flagix
	fa.Flag = flag
	fa.Claimer = flag.ConePos.Winner()
	if flag.IsWon {
		fa.IsClaimed = true
	} else {
		fa.OppFormationSize = flag.PlayerFormationSize(oppix)
		if flag.PlayerFormationSize(oppix) == flag.FormationSize() {
			fa.IsTargetMade = true
		}
		oppMissNo := flag.FormationSize() - flag.PlayerFormationSize(oppix)
		oppFlagStr := MoraleTroopsSum(flag.Players[oppix].Troops, flag.Players[oppix].Morales)
		oppMaxSum, _ := deckHand.Sum(oppix, card.COLNone, oppMissNo)
		targetBattStr := battStr(deckHand, oppix, oppMissNo, flag.Players[oppix].Troops, flag.Players[oppix].Morales)
		targetHostStr := oppMaxSum + oppFlagStr

		if len(flag.Players[oppix].Troops) > 0 {
			fa.TargetRank = CalcMaxRank(flag.Players[oppix].Troops,
				flag.Players[oppix].Morales, deckHand, oppix,
				flag.FormationSize(), flag.IsFog, combi.RankHost(flag.FormationSize()), targetHostStr, targetBattStr)
		} else {
			fa.TargetRank = calcMaxRankNewFlag(flag.Players[oppix].Morales, deckHand,
				oppix, flag.FormationSize(), flag.IsFog)
		}
		if fa.IsTargetMade {
			targetHostStr = targetHostStr + 1
			if targetBattStr != 0 {
				targetBattStr = targetBattStr + 1
			}
		}
		fa.TargetSum = targetHostStr
		fa.IsFog = flag.IsFog
		fa.FormationSize = flag.FormationSize()
		fa.BotFormationSize = flag.PlayerFormationSize(playix)
		if fa.BotFormationSize < flag.FormationSize() {
			fa.IsPlayable = true // this is only morale or troop
		}
		botMoraleNo := len(flag.Players[playix].Morales)
		if fa.BotFormationSize == botMoraleNo {
			fa.IsNewFlag = true //TODO what if moral 8 or 1,2,3 is alone then handAna is of no use.
			// It could only happen if moral card was played as second card and traitor need the
			// troop.
			fa.BotMaxRank = calcMaxRankNewFlag(flag.Players[playix].Morales, deckHand, playix,
				flag.FormationSize(), flag.IsFog)
		} else {
			fa.RankAnas = rankAnalyze(flag.Players[playix].Troops, flag.Players[playix].Morales, deckHand, playix,
				flag.FormationSize(), flag.IsFog, fa.TargetRank, fa.TargetSum, targetBattStr)
			fa.BotMaxRank = calcBotMaxRank(fa.RankAnas)
		}
		botMissNo := flag.FormationSize() - flag.PlayerFormationSize(playix)
		maxHostStr, _ := deckHand.Sum(playix, card.COLNone, botMissNo)
		botFlgStr := MoraleTroopsSum(flag.Players[playix].Troops, flag.Players[playix].Morales)
		fa.BotMaxSum = maxHostStr + botFlgStr
		botBattStr := battStr(deckHand, playix, botMissNo, flag.Players[playix].Troops, flag.Players[playix].Morales)
		if fa.IsTargetMade {
			botRankStr := fa.BotMaxSum
			if fa.BotMaxRank != combi.RankHost(fa.FormationSize) {
				botRankStr = botBattStr
			}
			//TargetSum and targetBattStr is the same for a made combination only rank differ
			fa.IsLost = lost(fa.TargetRank, fa.TargetSum, botRankStr, fa.BotMaxRank, fa.FormationSize)
		}
		if !fa.IsLost {
			fa.IsWin = isWin(fa.TargetRank, fa.IsTargetMade, fa.RankAnas, fa.FormationSize)
		}
		log.Printf(log.Debug, "Flagix:%v Flag:%v Lost:%v BotRank:%v BotSum:%v TargetRank:%v TargetSum:%v\n", fa.Flagix, fa.Flag, fa.IsLost, fa.BotMaxRank, fa.BotMaxSum, fa.TargetRank, fa.TargetSum)
	}

	return fa
}
func battStr(deckHandTroops *dht.Cache, playix, missNo int, flagTroops []card.Troop, flagMorales []card.Morale) (maxStr int) {
	if len(flagTroops) > 0 {
		flagColor := combi.FlagColor(flagTroops)
		if flagColor != card.COLNone {
			str, isOk := deckHandTroops.Sum(playix, flagColor, missNo)
			if isOk {
				maxStr = MoraleTroopsSum(flagTroops, flagMorales) + str
			}
		}
	} else {
		//maxColorix := card.COLNone
		if missNo > 0 {
			for colorix := 1; colorix < card.NOColors+1; colorix++ {
				str, isOk := deckHandTroops.Sum(playix, colorix, missNo)
				if isOk && str > maxStr {
					//		maxColorix = colorix
					maxStr = str
				}
			}
			if maxStr != 0 {
				maxStr = maxStr + MoraleTroopsSum(flagTroops, flagMorales)
			}
		} else {
			maxStr = MoraleTroopsSum(flagTroops, flagMorales)
		}
	}
	return maxStr
}
func calcBotMaxRank(combisAna []*combi.Analysis) (rank int) {
	for _, ana := range combisAna {
		if ana.Prop > 0 || ana.Comb.Formation.Value == card.FHost.Value {
			rank = ana.Comb.Rank
			break
		}
	}
	return rank
}

func isWin(
	anaTargetRank int,
	isTargetMade bool,
	combAnas []*combi.Analysis,
	formationSize int) (isWin bool) {

	for _, combiAna := range combAnas {
		if combiAna.Prop == 1 {
			if (!isTargetMade && combiAna.Comb.Rank <= anaTargetRank) ||
				(isTargetMade && combiAna.Comb.Rank < anaTargetRank) ||
				(isTargetMade && combiAna.Comb.Rank == anaTargetRank && combi.RankTieBreaker(anaTargetRank, formationSize).IsStrenght()) {
				isWin = true
			}
			break
		}
		if combiAna.Comb.Rank > anaTargetRank {
			break
		}
	}
	return isWin
}
func opp(ix int) (oppix int) {
	oppix = ix + 1
	if oppix > 1 {
		oppix = 0
	}
	return oppix
}

// HandAnalyze create a rank analysis of each troop on the hand.
// it assume no/ ignore enviroment cards,
// and targetSum = 0 this means can use Prob or Playables on combination Host
func HandAnalyze(
	deckHandTroops *dht.Cache,
	botix int,
) (ana map[card.Troop][]*combi.Analysis) {
	ana = make(map[card.Troop][]*combi.Analysis)
	targetRank := 1
	targetHostStr := 0
	targetBattStr := 0
	handTroops := deckHandTroops.SrcHandTroops[botix]
	for _, troop := range handTroops {
		simDeckHandTroops := deckHandTroops.CopyWithOutHand(troop, botix)
		ana[troop] = rankAnalyze([]card.Troop{troop}, nil, simDeckHandTroops, botix, 3, false, targetRank, targetHostStr, targetBattStr)
	}
	return ana
}

//lost calculate if a flag is lost it assume opponent moves first and win
//when rank or sum is equal
func lost(targetRank, targetStr, botStr, botRank, formationSize int) bool {
	isLost := false

	if targetRank < botRank || (targetRank == botRank && combi.RankTieBreaker(targetRank, formationSize).IsRank()) {
		isLost = true
	} else if targetRank == botRank &&
		combi.RankTieBreaker(targetRank, formationSize).IsStrenght() &&
		botStr < targetStr { //target Sum has been increased with one because it was made
		isLost = true
	}

	return isLost

}

// CalcMaxRank calculate the maximum rank a
// play may reach on a flag.
func CalcMaxRank(
	flagTroops []card.Troop,
	flagMorales []card.Morale,
	deckHandTroops *dht.Cache,
	playix int,
	formationSize int,
	isFog bool,
	targetRank, targetHostStr, targetBattStr int,
) (rank int) {

	combinations := combi.Combinations(formationSize)
	allCombi := math.Comb(uint64(len(deckHandTroops.SrcDeckTroops)), uint64(deckHandTroops.SrcDrawNos[playix]))
	for _, comb := range combinations {
		ana := combi.Ana(comb, flagTroops, flagMorales, deckHandTroops, playix, formationSize, isFog, targetRank, targetHostStr, targetBattStr)
		ana.SetAll(allCombi)
		if ana.Prop > 0 || comb.Formation.Value == card.FHost.Value {
			rank = ana.Comb.Rank
			break
		}
	}
	return rank
}
func calcMaxRankNewFlag(
	flagMorales []card.Morale,
	deckHandTroops *dht.Cache,
	playix int,
	formationSize int,
	isFog bool,
) (rank int) {
	if isFog {
		rank = combi.RankHost(formationSize)
	} else {
		combinations := combi.Combinations(formationSize)
		moraleNo := len(flagMorales)
		if moraleNo == 0 {
			rank = calcMaxRankNewFlagZeroMoral(deckHandTroops, playix, formationSize, combinations)
		} else if moraleNo == formationSize {
			rank = findRank(card.FBattalion.Value, 0, combinations) //TODO
		} else {
			rank = calcMaxRankNewFlagMorales(flagMorales, deckHandTroops, playix, formationSize, combinations)
		}
	}
	return rank
}
func calcMaxRankNewFlagZeroMoral(
	deckHandTroops *dht.Cache,
	playix, formationSize int,
	combinations []*combi.Combination) (rank int) {
	set := deckHandTroops.Set(playix)
	for _, comb := range combinations {
		if comb.Formation.Value == card.FWedge.Value {
			for _, troops := range comb.Troops {
				made := true
				for _, troop := range troops {
					if !set[troop] {
						made = false
						break
					}
				}
				if made {
					rank = comb.Rank
					return rank
				}
			}
		} else {
			break
		}
	}
	sortStrs := deckHandTroops.SortStrs(playix)
	for str := len(sortStrs) - 1; str > 0; str-- {
		troops := sortStrs[str]
		if formationSize <= len(troops) {
			rank = findRank(card.FPhalanx.Value, str*formationSize, combinations)
			return rank
		}
	}
	isBatt := false
	for color := 1; color <= card.NOColors; color++ {
		_, isOk := deckHandTroops.Sum(playix, color, formationSize)
		if isOk {
			isBatt = true
			break
		}
	}
	if isBatt {
		rank = findRank(card.FBattalion.Value, 0, combinations)
		return rank
	}
	inRow := 0
	for str := len(sortStrs) - 1; str > 0; str-- {
		troops := sortStrs[str]
		if len(troops) > 0 {
			inRow = inRow + 1
			if inRow == formationSize {
				sum := formationSize * (formationSize + 1) / 2
				sum = sum + (str-1)*formationSize
				rank = findRank(card.FSkirmish.Value, sum, combinations)
				return rank
			}
		} else {
			inRow = 0
		}
	}
	return combi.RankHost(formationSize)
}

func calcMaxRankNewFlagMorales(
	morales []card.Morale,
	deckHandTroops *dht.Cache,
	playix int,
	formationSize int,
	combinations []*combi.Combination) (rank int) {
	set := deckHandTroops.Set(playix)
	jockers := newJockers(morales)
	for _, comb := range combinations {
		if comb.Formation.Value == card.FWedge.Value {
			for _, troops := range comb.Troops {
				made := true
				for _, troop := range troops {
					if !set[troop] {
						if !jockers.use(troop.Strenght()) {
							made = false
							break
						}
					} else {
						jockers.validate(troop.Strenght())
					}
				}
				if made && jockers.confirm() {
					rank = comb.Rank
					return rank
				}
				jockers.reset()
			}
		} else {
			break
		}
	}
	for str, troops := range deckHandTroops.SortStrs(playix) {
		if formationSize-len(morales) <= len(troops) && jockers.validateAll(str) {
			rank = findRank(card.FPhalanx.Value, str*formationSize, combinations)
			return rank
		}
	}
	isBatt := false
	for color := 1; color <= card.NOColors; color++ {
		_, isOk := deckHandTroops.Sum(playix, color, formationSize-len(morales))
		if isOk {
			isBatt = true
		}
	}
	if isBatt { ///Cant handle len(morales)=formationSize
		rank = findRank(card.FBattalion.Value, 0, combinations)
		return rank
	}
	jockers.reset()
	sortStrs := deckHandTroops.SortStrs(playix)
	for startStr := len(sortStrs) - 1; startStr >= formationSize; startStr-- {
		for i := 0; i < formationSize; i++ {
			str := startStr - i
			troops := sortStrs[str]
			if len(troops) > 0 || jockers.use(str) {
				if len(troops) > 0 {
					jockers.validate(str)
				}
				if i == formationSize-1 {
					if jockers.confirm() {
						sum := formationSize * (formationSize + 1) / 2
						sum = sum + (str-1)*formationSize
						rank = findRank(card.FSkirmish.Value, sum, combinations)
						return rank
					}
				}
			} else {
				break
			}
		}
		jockers.reset()
	}
	return combi.RankHost(formationSize)
}

type jockers struct {
	isValids []bool
	morales  []card.Morale
	isUseds  []bool
}

func newJockers(morales []card.Morale) (j *jockers) {
	j = new(jockers)
	var leader card.Morale
	for _, morale := range morales {
		if !morale.IsLeader() {
			j.morales = append(j.morales, morale)
			j.isValids = append(j.isValids, false)
		} else {
			leader = morale
		}
		j.isUseds = append(j.isUseds, false)
	}
	if leader != 0 {
		j.morales = append(j.morales, leader)
		j.isUseds = append(j.isUseds, false)
	}
	return j
}
func (j *jockers) use(str int) (isOk bool) {
	for i, morale := range j.morales {
		if !isOk && !j.isUseds[i] && morale.ValidStrenght(str) {
			j.isUseds[i] = true
			isOk = true
		}
		if !morale.IsLeader() && !j.isValids[i] && morale.ValidStrenght(str) {
			j.isValids[i] = true
		}
	}
	return isOk
}
func (j *jockers) confirm() (isOk bool) {
	isOk = true
	for i, morale := range j.morales {
		if !morale.IsLeader() && !j.isValids[i] {
			isOk = false
			break
		}
	}
	return isOk
}
func (j *jockers) validate(str int) {
	for i, morale := range j.morales {
		if !morale.IsLeader() && !j.isValids[i] && morale.ValidStrenght(str) {
			j.isValids[i] = true
		}
	}
}
func (j *jockers) validateAll(str int) bool {
	for _, morale := range j.morales {
		if !morale.IsLeader() && !morale.ValidStrenght(str) {
			return false
		}
	}
	return true
}
func (j *jockers) reset() {
	for i, morale := range j.morales {
		j.isUseds[i] = false
		if !morale.IsLeader() {
			j.isValids[i] = false
		}
	}
}
func (j *jockers) maxStr() (str int) {
	for _, morale := range j.morales {
		str = str + morale.MaxStrenght()
	}
	return str
}

func findRank(formationValue, strenght int, combinations []*combi.Combination) (rank int) {
	for _, v := range combinations {
		if v.Strength == strenght && v.Formation.Value == formationValue {
			rank = v.Rank
			break
		}
	}
	if rank == 0 {
		panic(fmt.Sprintf("Could not find rank for formation value %v, strenght:%v", formationValue, strenght))
	}
	return rank
}

//MoraleTroopsSum sums the strenght of troops and morales.
//Morales use the maxStrenght.
func MoraleTroopsSum(flagTroops []card.Troop, flagMorales []card.Morale) (sum int) {
	for _, troop := range flagTroops {
		sum = sum + troop.Strenght()
	}
	for _, morale := range flagMorales {
		sum = sum + morale.MaxStrenght()
	}
	return sum
}

func rankAnalyze(
	flagTroops []card.Troop,
	flagMorales []card.Morale,
	deckHandTroops *dht.Cache,
	playix int,
	formationSize int,
	isFog bool,
	targetRank, targetHostStr, targetBattStr int,
) (ranks []*combi.Analysis) {
	combinations := combi.Combinations(formationSize)
	formationMade := len(flagTroops)+len(flagMorales) == formationSize
	ranks = make([]*combi.Analysis, len(combinations))
	allCombi := math.Comb(uint64(len(deckHandTroops.SrcDeckTroops)), uint64(deckHandTroops.SrcDrawNos[playix]))
	for i, comb := range combinations {
		ana := combi.Ana(comb, flagTroops, flagMorales, deckHandTroops, playix, formationSize, isFog, targetRank, targetHostStr, targetBattStr)
		ana.SetAll(allCombi)
		ranks[i] = ana
		if formationMade && ana.Prop == 1 {
			break
		}
		//TODO make prob and playables for host, and include isFog in rankAnalyze.
		//Slight change to isWin(): == for host possible. and the isLost can reuse the rankAnalyze but
		//it must also work without.First make without prob calc just set .5 for undesided.
	}
	return ranks
}
