package model

type News struct {
	ID      int     `json:"id"`
	Title   string  `json:"title"`
	Content *string `json:"content"`
	UserID  int     `json:"user_id"`
}

func (m News) TableName() string {
	return "news"
}
