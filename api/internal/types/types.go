package types

type User struct {
	UserID    int64   `json:"user_id" db:"user_id"` // ID in table of interns
	Phone     string  `json:"phone" db:"phone"`
	Verified  bool    `json:"verified" db:"verified"`   // Is user verified
	Email     string  `json:"email" db:"email"`         // User email
	Course    *string `json:"course" db:"course"`       // ID of selected course
	PersonID  *string `json:"person_id" db:"person_id"` // ID of user profile in Notion
	ProfileID *string `json:"profile_id" db:"profile_id"`
	FIO       string  `json:"fio" db:"fio"`
	Username  string  `json:"username" db:"username"`
	Avatar    string  `json:"avatar" db:"avatar"`
	State     int     `json:"-" db:"state"`
	Fails     int     `json:"-" db:"fails"`
}

const (
	TaskStatusNotStarted = iota + 1
	TaskStatusInProgress
	TaskStatusChecking
	TaskStatusDone
	TaskStatusRejected
)

var TaskStatuses = map[int]string{
	TaskStatusNotStarted: "Не начато",
	TaskStatusInProgress: "В процессе",
	TaskStatusChecking:   "На проверке",
	TaskStatusDone:       "Выполнена",
}

type Course struct {
	Name      string `json:"name" db:"name"`
	NotionID  string `json:"notion_id" db:"course_id"`
	Invite    string `json:"invite" db:"invite_link"`
	CuratorID int64  `json:"curator_id" db:"curator_id"`
	ShortName string `json:"short_name" db:"short_name"`
	GroupID   int64  `json:"group_id" db:"group_id"`
}

type Task struct {
	TaskID      string  `json:"task_id" db:"task_id"`
	UserID      int64   `json:"-" db:"user_id"`
	Title       string  `json:"title" db:"title"`
	Status      int     `json:"-" db:"status"`
	Content     string  `json:"content,omitempty" db:"content"`
	CourseID    string  `json:"course_id,omitempty" db:"course_id"`
	Next        *string `json:"next,omitempty" db:"next"`
	CompletedAt int64   `json:"completed_at,omitempty" db:"completed_at"`
	Homework    *string `json:"homework,omitempty" db:"homework"`
	IsFirst     bool    `json:"-" db:"is_first"`
	Cover       string  `json:"cover,omitempty" db:"cover"`
	Section     string  `json:"section,omitempty" db:"section"`
	Type        string  `json:"type" db:"type"`

	StatusPublic string `json:"status" db:"-"`
}
