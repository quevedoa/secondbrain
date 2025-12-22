package entity

type Note struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type Query struct {
	Text       string `json:"text"`
	NumResults int32  `json:"numResults"`
}
