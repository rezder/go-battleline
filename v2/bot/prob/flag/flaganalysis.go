package flag

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob/combi"
	"github.com/rezder/go-battleline/v2/bot/prob/dht"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
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
			fa.TargetRank, _ = CalcMaxRank(flag.Players[oppix].Troops,
				flag.Players[oppix].Morales, deckHand, oppix,
				flag.FormationSize(), flag.IsFog, combi.RankHost(flag.FormationSize()), targetHostStr, targetBattStr)
		} else {
			fa.TargetRank = calcMaxRankNewFlag(flag.Players[oppix].Morales, deckHand,
				oppix, flag.FormationSize(), targetBattStr, flag.IsFog)
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
				flag.FormationSize(), targetBattStr, flag.IsFog)
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
) (rank int, prob float64) {

	combinations := combi.Combinations(formationSize)
	allCombi := math.Comb(uint64(len(deckHandTroops.SrcDeckTroops)), uint64(deckHandTroops.SrcDrawNos[playix]))
	for _, comb := range combinations {
		ana := combi.Ana(comb, flagTroops, flagMorales, deckHandTroops, playix, formationSize, isFog, targetRank, targetHostStr, targetBattStr)
		ana.SetAll(allCombi)
		if ana.Prop > 0 || comb.Formation.Value == card.FHost.Value {
			rank = ana.Comb.Rank
			prob = ana.Prop
			break
		}
	}
	return rank, prob
}
func calcMaxRankNewFlag(
	flagMorales []card.Morale,
	deckHandTroops *dht.Cache,
	playix, formationSize, targetBattStr int,
	isFog bool,
) (rank int) {
	if isFog {
		rank = combi.RankHost(formationSize)
	} else {
		combinations := combi.Combinations(formationSize)
		moraleNo := len(flagMorales)
		if moraleNo == 0 {
			rank = calcMaxRankNewFlagZeroMoral(deckHandTroops, playix, formationSize, targetBattStr, combinations)
		} else if moraleNo == formationSize {
			str := MoraleTroopsSum(nil, flagMorales)
			rank = findRank(card.FBattalion.Value, str, combinations, targetBattStr)
		} else {
			rank = calcMaxRankNewFlagMorales(flagMorales, deckHandTroops, playix, formationSize, targetBattStr, combinations)
		}
	}
	return rank
}
func calcMaxRankNewFlagZeroMoral(
	deckHandTroops *dht.Cache,
	playix, formationSize, targetBattStr int,
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
			rank = findRank(card.FPhalanx.Value, str*formationSize, combinations, targetBattStr)
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
		str := battStr(deckHandTroops, playix, formationSize, nil, nil)
		rank = findRank(card.FBattalion.Value, str, combinations, targetBattStr)
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
				rank = findRank(card.FSkirmish.Value, sum, combinations, targetBattStr)
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
	playix, formationSize, targetBattStr int,
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
			rank = findRank(card.FPhalanx.Value, str*formationSize, combinations, targetBattStr)
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
		str := battStr(deckHandTroops, playix, formationSize-len(morales), nil, morales)
		rank = findRank(card.FBattalion.Value, str, combinations, targetBattStr)
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
						rank = findRank(card.FSkirmish.Value, sum, combinations, targetBattStr)
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

type rankProb struct {
	rank int
	prob float64
}

func (rp rankProb) String() string {
	return fmt.Sprintf("Rank{rank:%v,prob:%v}", rp.rank, rp.prob)
}
func (rp rankProb) mf() string {
	return fmt.Sprintf("%v,%.3f", rp.rank, rp.prob)
}

//TODO maybe add sum as batt and host get prob = 1 for opp

//TfAna tensor flow analysis
type TfAna struct {
	flagix       int
	conePos      pos.Cone
	oppIsNewFlag bool
	oppMissNo    int
	oppRank      rankProb
	botIsNewFlag bool
	botMissNo    int
	botRanks     [5]rankProb
}

func (tfa *TfAna) String() (txt string) {
	return fmt.Sprintf("TfAna{flag(ix:%v,pos:%v),Opp(isNew:%v,missNo:%v,rank:%v},Bot(IsNew:%v,missNo:%v,ranks:%v)",
		tfa.flagix, tfa.conePos, tfa.oppIsNewFlag, tfa.oppMissNo, tfa.oppRank, tfa.botIsNewFlag, tfa.botMissNo, tfa.botRanks)
}

// MachineFormat creates a tensor flow string.
func (tfa *TfAna) MachineFormat() (txt string) {
	txt = fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v",
		mfConvConePos(tfa.conePos), mfConvBool(tfa.oppIsNewFlag), tfa.oppMissNo, tfa.oppRank.mf(), mfConvBool(tfa.botIsNewFlag), tfa.botMissNo,
		tfa.botRanks[0].mf(), tfa.botRanks[1].mf(), tfa.botRanks[2].mf(), tfa.botRanks[3].mf(), tfa.botRanks[4].mf())
	return txt
}
func (tfa *TfAna) MachineFloats() (floats [19]float32) {
	if tfa != nil {
		for i, f := range mFloatConvConePos(tfa.conePos) {
			floats[i] = f
		}
		floats[3] = mFloatsConvBool(tfa.oppIsNewFlag)
		floats[4] = float32(tfa.oppMissNo)
		floats[5] = float32(tfa.oppRank.rank)
		floats[6] = float32(tfa.oppRank.prob)
		floats[7] = mFloatsConvBool(tfa.botIsNewFlag)
		floats[8] = float32(tfa.botMissNo)
		for i, r := range tfa.botRanks {
			floats[9+i*2] = float32(r.rank)
			floats[10+i*2] = float32(r.prob)
		}
	}
	return floats
}

//NewTfAnalysis create a tensor flow flag analysis.
func NewTfAnalysis(
	playix int,
	flag *game.Flag,
	deckHandTroops *dht.Cache,
	flagix int) (tfa *TfAna) {
	tfa = new(TfAna)
	oppix := opp(playix)
	tfa.flagix = flagix
	tfa.conePos = flag.ConePos
	if !flag.IsWon {
		oppHostStr := 0
		oppBattStr := 0
		if flag.HasFormation(oppix) {
			oppFormation, oppStr := flag.Formation(oppix)
			oppRank := findRank(oppFormation.Value, oppStr, combi.CombinationsMud(flag.IsMud), oppStr)
			tfa.oppRank = rankProb{rank: oppRank, prob: 1}
			oppHostStr = oppStr + 1
			if oppFormation.Value == card.FBattalion.Value { //TODO maybe also wedge
				oppBattStr = oppStr + 1
			}
		} else {
			tfa.oppMissNo = flag.FormationSize() - flag.PlayerFormationSize(oppix)
			oppFlagStr := MoraleTroopsSum(flag.Players[oppix].Troops, flag.Players[oppix].Morales)
			oppMaxSum, _ := deckHandTroops.Sum(oppix, card.COLNone, tfa.oppMissNo)
			oppHostStr = oppMaxSum + oppFlagStr
			oppBattStr = battStr(deckHandTroops, oppix, tfa.oppMissNo, flag.Players[oppix].Troops, flag.Players[oppix].Morales)

			if len(flag.Players[oppix].Troops) == 0 {
				oppRank := calcMaxRankNewFlag(flag.Players[oppix].Morales, deckHandTroops, oppix, flag.FormationSize(), oppBattStr, flag.IsFog)
				tfa.oppRank = rankProb{rank: oppRank, prob: 0.5} //TODO we do not have a prob what to do 0,0.5 or 1 and add field is new
				tfa.oppIsNewFlag = true
			} else {
				oppRank, oppProb := CalcMaxRank(flag.Players[oppix].Troops,
					flag.Players[oppix].Morales, deckHandTroops, oppix,
					flag.FormationSize(), flag.IsFog, combi.RankHost(flag.FormationSize()), oppHostStr, oppBattStr)
				tfa.oppRank = rankProb{rank: oppRank, prob: oppProb}
			}

		}
		if flag.HasFormation(playix) {
			formation, str := flag.Formation(playix)
			rank := findRank(formation.Value, str, combi.CombinationsMud(flag.IsMud), oppBattStr)
			tfa.botRanks[0] = rankProb{rank: rank, prob: 1}
		} else {
			tfa.botMissNo = flag.FormationSize() - flag.PlayerFormationSize(playix)
			if len(flag.Players[playix].Troops) == 0 {
				botRank := calcMaxRankNewFlag(flag.Players[playix].Morales, deckHandTroops, playix, flag.FormationSize(), oppBattStr, flag.IsFog)
				tfa.botRanks[0] = rankProb{rank: botRank, prob: 1} //TODO we do not have a prob what to do 0,0.5 or 1
				tfa.botIsNewFlag = true
			} else {
				rankAnas := rankAnalyze(flag.Players[playix].Troops, flag.Players[playix].Morales, deckHandTroops, playix,
					flag.FormationSize(), flag.IsFog, tfa.oppRank.rank, oppHostStr, oppBattStr)
				i := 0
				for _, combiAna := range rankAnas {
					if combiAna.Prop > 0 {
						tfa.botRanks[i] = rankProb{rank: combiAna.Comb.Rank, prob: combiAna.Prop}
						i++
						if i == len(tfa.botRanks)-1 {
							break
						}
					}
				}

			}
		}
	}
	return tfa
}

func mfConvBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
func mFloatsConvBool(b bool) float32 {
	if b {
		return 1
	}
	return 0
}
func mfConvConePos(c pos.Cone) (txt string) {
	switch c {
	case pos.ConeAll.Players[0]:
		return "0,1,0"
	case pos.ConeAll.Players[1]:
		return "0,0,1"
	}
	return "1,0,0"
}

func mFloatConvConePos(c pos.Cone) (floats [3]float32) {
	switch c {
	case pos.ConeAll.Players[0]:
		return [3]float32{0, 1, 0}
	case pos.ConeAll.Players[1]:
		return [3]float32{0, 0, 1}
	}
	return [3]float32{1, 0, 0}
}
func findRank(formationValue, strenght int, combinations []*combi.Combination, targetBattStr int) (rank int) {
	srcStr := strenght
	if formationValue == card.FBattalion.Value || formationValue == card.FHost.Value {
		srcStr = 0
	}
	for _, v := range combinations {
		if v.Strength == srcStr && v.Formation.Value == formationValue {
			rank = v.Rank
			break
		}
	}
	if formationValue == card.FBattalion.Value {
		if targetBattStr == 0 {
			rank = rank + 1
		} else {
			if strenght < targetBattStr { //assume equal str is ennough (oppBattStr +1 for made if needed)
				rank = rank + 1
			}
		}
	}
	if rank == 0 {
		panic(fmt.Sprintf("Could not find rank for formation value %v, strenght:%v", formationValue, strenght))
	}
	return rank
}
