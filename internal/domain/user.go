package domain

type User struct {
	ID       int64
	Email    string
	Password string
	CTime    int64
	UTime    int64

	NickName    string
	BirthDate   string
	Description string
}
