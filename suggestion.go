package ns

type Suggestion struct {
	Display string
	Value   string
}

type Suggestions struct {
	Total int // reflects the total (could be longer than len(.Items)
	Items []*Suggestion
}

func NewSuggestions() *Suggestions {
	return &Suggestions{
		Items: []*Suggestion{},
	}
}

// Add will always reset the total after the suggestion is added, thus it is necessary to set the total
// after, if it is different than the length of suggestions
func (s *Suggestions) Add(suggestion *Suggestion) {
	s.Items = append(s.Items, suggestion)
	s.Total = len(s.Items)
}

func NewSuggestion(display string, value string) *Suggestion {
	return &Suggestion{
		Display: display,
		Value:   value,
	}
}
