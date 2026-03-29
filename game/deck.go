package game

import "math/rand/v2"

// ShuffleDeck randomises the order of a player's deck in place.
func ShuffleDeck(p *Player) {
	rand.Shuffle(len(p.Deck), func(i, j int) {
		p.Deck[i], p.Deck[j] = p.Deck[j], p.Deck[i]
	})
}

// DrawCards moves up to n cards from the top of the player's deck into their hand.
// Returns the number of cards actually drawn (may be less if the deck is exhausted).
func DrawCards(p *Player, n int) int {
	drawn := 0
	for i := 0; i < n && len(p.Deck) > 0; i++ {
		card := p.Deck[0]
		p.Deck = p.Deck[1:]
		p.Hand = append(p.Hand, card)
		drawn++
	}
	return drawn
}
