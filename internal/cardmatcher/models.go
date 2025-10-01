package cardmatcher

type Card struct {
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
	Legality string `json:"legality"`
}

type CardMatch struct {
	Card       Card
	Similarity float64
	Distance   int
}

type CardDatabase struct {
	cards []Card
}
