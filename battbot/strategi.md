# Strategy for bot.

Calculate the probability for every combination on a flag for both players. A combination is a formation and a sum. For every combination find the cards of that combination in the hand.

## The strategies for a flag.

1. strategy a combination that is higher than any from the opponent and have a higher probability.

2. strategy a combination that is higher and lower probability but close.

3. strategy a combination that is equal and a probability that is higher or close. For both strategies 2 and 3 look at the next lower formation for one that is better and with a probability that better or close if one exist and have a commen card with the higher formation play the strategy.

4. strategy the combination with highest probability.

### The strategy for which card if more exist for a combination.
 The higest except if you have a whole top combination then the lowest card.

## The strategy for which flag to play.

Every flag has a value depending on the place and distance to a lost flag.

### Place.
* Flag 1 and 9 value 1
* Flag 2 and 8 value 2
* Flag 3 to 7 value 3

### Distance to lost flag.
* Distance 1 value 4
* Distance 2 value 3

We may play around with the values but it is clear that a flag next to two lost flag have the higest priority.

Every flag have a speed that is priority to finish fast.

1. Lost flag.
2. Equal top formation.
3. More cards played on a flag by a opponent. More cards is the cards of the opponent - cards of the bot.

## Math model.
The model view the problem as the properbility of making the combination given you get half the deck of cards. The it is just n over x problems.

Numbers of unknown cards: n 
Numbers of cards to draw: d
Numbers of cards that can be used in the combination: v
Numbers of cards that can not be used in the combination: y
Numbers of cards in the combination: c
Numbers of cards that must be used: mc

Notation (n/x) = !n/(!x * !(n-x)) n over x. Hopefully it will not conflict with division.

Numbers of draws with v:

(v/mc) * (y/(d-mc)) + (v/mc+1) * (y/(d-(mc+1)) +...+ (v/v) * (y/(d-v)) 

Numbers of all possible draws.

(n/d)

The v draws may not all be valid so the (v/x) must reduced with non valid combinations.
Before the numbers can be divided and reveal the properbility of a combination.

* n is equal to the remaining cards in the deck together with the cards on the oppent hand.
* d is equal to the remaining cards in deck divided by 2 and rounded up for the player and down for the opponent as player draw first. The opponent also get is cards on the hand.
* mc >= c for c+y >= d
* mc = d-y for c+y < d
* y = n-v

### Example The first and biggest when player starts.
Deck: 46

Player:

n = 46 + 7 = 53
d = 46/2 = 23 

Opponent:

n = 46 + 7 = 53
d = 46/2 + 7 = 30

This is big numbers !53 is not a valid 64 bit integer. So we need special math.
uint64 is 18446744073709551615 maybe we only need to do some reductions.

(53/23) = 115061434509375 if I trust my own paper calculations.

Reduction is unpredictable with regards to overflow as some problems will reduce better than other independed of the size.

The max valid of v:

* Wedge: v = c = 4
* Phalanx: v = 6
* Battalion Order: v = 10
* Skirmish Line: v = 24 
* Host: v = n

This is even if the formation is restricted to a combination for example Battalion Order with the strenght of 13 make use of all cards.

### Example Skirmish Line.
Player have a 6 on the flag and two eights and two sevens remains in the deck or on the opponent hand.

Deck: 20

Player:

n = 26
d = 10
v = 4, 2 eights and 2 sevens.
c = mc = 2
y = 26-4 = 22

**Numbers of draws with v:**

2 cards:

(v/mc) = (4/2) = 6 but two combinations is not valid 2 eights and two sevens.
(y/d-mc) = (22/8) = 11 * 5 * 19 * 3 * 17 * 2 * 3 = 11 * 19 * 17 * 90 = 319.770
4 * 319770

3 cards:

(v/mc+1) = (4/3) = 4 all valid.
(y/d-(mc+1)) = (22/7) = 11 * 19 * 3 * 17 * 16 = 170.544
4 * 170544

4 cards:

(v/v) = (4/4) = 1
(y/(d-v)) = (22/6) = 11 * 21 * 19 * 17 = 74613
1 * 74613

Valid combinations 4 * 319770 + 4 * 170544 + 74613 = 1279080 + 682176 + 74613 = 2.035.869

**Numbers of all possible draws.**

(n/d) = (26/10)= 26 * 25 * 23 * 22 * 19 * 17 / 20 =  5.311.735

2035869 / 5311735 = 0.383

For control:

1 cards:

(4/1) = 4 
(22/9) = 22 * 19 * 17 * 15 * 14 / 3 = 497.420
4 * 497420 = 1.989.680

0 cards:

(4/0) = 1
(22/10) = 22 * 19 * 17 * 14 * 13 / 2 = 646.646

646.646 + 1.989.680 + 1.918.620 + 682.176 + 74613 = 5.311.735

## Functions
The best way is properly calculate two ordered list of Combination.
Where a Combination is a struct.
```
type Combination struct {
    rank int
    formation int
    strenght int
    validCards map[int][]int
}
```
ValidCard is the valid cards per color. We will need a none color for formation without color. We need two list because of the mud.
Maybe the rank should not be in the combination.

Then use the lists for all calculation which then will be mostly about searching the validCards.
Keep the list as static data no change of slices, copy the slices to a map/set for easy serach and manipulations.













