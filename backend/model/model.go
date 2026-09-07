package model

type Post struct {
	Id        string   `json:"id"`
	User      string   `json:"user"`
	Message   string   `json:"message"`
	Url       string   `json:"url"`
	Type      string   `json:"type"`
	CreatedAt int64    `json:"createdAt,omitempty"`
	UpdatedAt int64    `json:"updatedAt,omitempty"`
	Likes     []string `json:"likes,omitempty"`
}
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Age      int64  `json:"age"`
	Gender   string `json:"gender"`
}
type Session struct {
	Username  string `json:"username"`
	ExpiresAt int64  `json:"expiresAt"`
}
