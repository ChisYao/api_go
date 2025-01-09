package domain

type User struct {
	Id       int64
	Email    string
	Password string
	Name     string
	Phone    string
	Gender   string
	//CTime time.Time
}
