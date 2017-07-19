# Machine learning bot

The plan is to see the game as independed move based on position, I think that is ok because
the order of move really does not matter. So for machine learning to work the game position 
must formulated to fit machine learning. I feel a little stupid realising that the game position
can be viewed as 70 cards and their positions and 9 flags and their position and any move is just
a change of postions. The game position as view from the bot also need the numbers of tactic
and troop cards of the opponent to have a complete picture.
Basicly every card game be described a number of cards and their positions and a move as a number
of cards to move and their new positions.

## Game postion

### Variables cards

card1-card70 containing a position.

Card Position: 

* **Flag** opponent/bot denoted FLAGxO,FLAGxB  x is a number from 1-9. 
* **Dish** DISHOpp or DISHBot
* **Hand** HANDLegal or HAND Legal the cards that can be moved.
* **Deck** DECK 

This gives 18+2+3=23 positions.

### Varaibles scout return
* **SReturnBot1** First returned card 1-70 excepet 64 the scout return and 0 for no card.
cleared when card hit the table.
* **SReturnBot2** Second returned card 1-70 excepet 64 the scout return and 0 for no card.
cleared when card hit the table
* **OppKnownDeckTacNo** The number of cards the opponent knows in tactic deck
* **OppKnownDeckTroopNo** The number of cards the opponent knows in troop deck
* **OppKnownHand1** The opponent may know up till two cards as a result of the opponent haveing played scout return.
* **OppKnownHand2** See above and these fields is cleared when **both** cards have hit the table.
The have information about the known deck numbers of cards. We could clear if know dbuckStd     = "Std" // Card to flag and pass.
	buckDeck    = "Deck"
	buckClaim   = "Claim"
	buckSpecial = "Special" //Traitor,Scout,Deserter, Redeploy, scoutReturn.
	buckGames   = "Game"
	buckWin     = "Win"
	buckLose    = "Lose"
	keyMeta     = "Meta"eck cards.

### Variables flags

flag1-flag9 containing a position.

Flags position: 

* **Unclaimed** CLAIMNon
* **Claimed**  CLAIMBot,CLAIMOpp 

To the old code CLAIMNon=0,CLAIMBot=1,CLAIMOpp=-1 this is not
good if I want to use unit8(byte) changing CLAIMOpp=2.

### Variables opponent hand.

troopNo,tacNo containing a number between 0-7.
The opponent can only have maximum 7 cards when it is the bot turn. So both numbers is
between 0-7.

* **Troop cards** TROOPNo the number of troops including any known from scout returned.
* **Tactic cards** TACNo the number of troops including any known from scout returned.

### Varaiables pass

* **Pass** One if pass is allowed 0 else.
## Move

Every move need learn on it own but based on the same data except claim 
flag that is different

### Hand card move

The hand move be genralized to two cards and a destination flag/dish/deck.
The first card is card from hand second is card from flag.

* **Redeploy** first card is redeploy second card is the card that is moved
and the flag/dish is the new position of the second card.

* **Scout** first and second card is none and position is deck tac or troop. 
The deck is not used for hand card move but we need it for other moves

Legal moves could be calculated for every first card.

* **Scout** that is easy as the deck select is not handle here. Move 1.
* **Redeploy** Second card can be all troops, enviroment and morale card 66 and
all flags and dish. Moves 660.
* **Deserter** can kill the same 66 cards as redeploy can move. Move 66.
* **Traitor** can only move troops 60 to a flag. Moves 540.
* **Troop, enviroment or morale** The 66 cards can be move to any 9 flag
594 and the pass move could be all none and dish. Moves 595.

This give 1+660+66+540+595=1862 hand card moves a lot but if the use that
we know the order of the cards then a none move could be could be desribed
as 1-7 the numbers of cards on the hand. That gives 7x9 moves 63. We could
do something simular for picking a card from a flag as we know which half
we get 9 flags time 4(3 for traiter) spots 36 not as big a reduction for
the other cards. 1+63+360+36+324=784 for none only 1862-595+63=1330.
It make more sense to use the first card as the hand card 1-7 and the
second as optional card 1-66 and desitination 0-9 zero is dish.
7x67x10=4590+1 or 7x36x10+1=2520+1.
Maybe we do not need on class that fit all maybe we only need y to have more
ones, so the problem have 0-7 notes first card 8-75 notes second card and
76-85 notes for destination and the sum of is one with in each group.



### Deck move

The deck move is based on the same position as hand card move.

The move is simple just troop or tactic so by adding it to the destination it
can be include in the hand card move.

### Claim move

The claim move can reduced to a single flag question with true and false and
the position can be reduced to 70 cards with less positions. FLAGB,FLAGO,VISUAL 
and NONVISUAL.
We still like to record the move in the same move as the other moves, but
9 flags claimed or not claimed equeal to 2^9=512, Adding "Claim" to
the first card and using bits to denote if a flag have been claimed.
we need 8 bits so both second and destination must be used.
```
func convertTo(x uint8)(flags [8]bool){
	for i := uint(0); i < 8; i++ {
		flags[i]=x & (1 << i) >> i==1
    }
    return flags
}
func convertFrom(flags [8]bool)(x uint8){
	for i,b := range flags{
		if b{
			x=x|(1<<uint8(i))
		}
    }
    return x
}
```

### Scout return move

Consits of two cards to record it we use the first and second card and
the Deck as desitnation. First is the first returned card. Cards is the hand
number.


## Definition

The whole position and move can be descrived with a simple 70+9+2+3 array
of uint8 which equal to byte.

### Card positions

CPDishBot=0
CPFlag1Bot=1
CPFlag2Bot=2
CPFlag3Bot=3
CPFlag4Bot=4
CPFlag5Bot=5
CPFlag6Bot=6
CPFlag7Bot=7
CPFlag8Bot=8
CPFlag8Bot=9
CPDishOpp=10
CPFlag1Opp=11
CPFlag2Opp=12
CPFlag3Opp=13
CPFlag4Opp=14
CPFlag5Opp=15
CPFlag6Opp=16
CPFlag7Opp=17
CPFlag8Opp=18
CPFlag9Opp=19
CPHandLegal=20
CPHand=21
CPDeck=22

### Flag positions

CLAIMNon=0
CLAIMBot=1
CLAIMOpp=2

### Move first card
Cardix 1-9 could use hand number hand number legal or card. I like card and modify when making data.
NoneCard=0
SPCClaimFlag=100

### Move second card
Cardix 1-70 ex 64 and 61,62,63 can only be in scout return
NoneCard=0
ClaimFlag for flag 9 one or zero one claimed

### Move destination

MDDish=0
MDFlag1=1
MDFlag2=2
MDFlag3=3
MDFlag4=4
MDFlag5=5
MDFlag6=6
MDFlag7=7
MDFlag8=8
MDFlag9=9
MDDeckTac=10
MDDeckTroop=11
MDDeck=12
ClaimFlag is 8 bits where first bit is flag1 up to flag8. 1 is claimed.


### Machine Learn Data

A array of 92 uint8 (byte)
0 Move index
1-70 Card postions in cardix order.
71-79 Flag positions int flagix order.
80 Opponent number of troops cards. 
81 Opponent number of tactic cards.
82 Scout return bot first card
83 Scout return bot second card
84 The number of cards the oppponent knows in tactic card deck.
85 The number of cards the oppponent knows in troop card deck.
86 Opponent know card on bot hand
87 Opponent know card on bot hand
88 Pass
89 Move hand card.
90 Move flag card.
91 Move destination.


## Database

The database could be a "Game" bucket with key from the game database
containing a "Detail" bucket.

The Detail bucket has keys Start NoStart MetaData

NoStart has buckets: Handmoves,DeckMoves, Claim moves
Start has buckets: --||--
MetaData contain aggregated information ex. The winner, giveup, drawn/played tactic cards starter/no starter, Number moves. Avg strengh of cards starter/nostarter 



