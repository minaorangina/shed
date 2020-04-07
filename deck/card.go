package deck

// Card represents a playing card
// The interface will allow for introducing a JokerCard in future
type Card interface {
	Rank() string
	Suit() string
	String() string
}
