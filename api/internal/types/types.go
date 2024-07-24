package types

type User struct {
	UserID    int64   `json:"id" db:"user_id"` // ID in table of interns
	Phone     string  `json:"phone" db:"phone"`
	Verified  bool    `json:"verified" db:"verified"`  // Is user verified
	Email     string  `json:"email" db:"email"`        // User email
	Course    *string `json:"course" db:"course"`      // ID of selected course
	PersonID  *string `json:"personID" db:"person_id"` // ID of user profile in Notion
	ProfileID *string `json:"profileID" db:"profile_id"`
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
	NotionID  string `json:"notionID" db:"course_id"`
	Invite    string `json:"invite" db:"invite_link"`
	CuratorID int64  `json:"curatorID" db:"curator_id"`
	ShortName string `json:"shortName" db:"short_name"`
	GroupID   int64  `json:"groupID" db:"group_id"`
}

type Task struct {
	TaskID      string  `json:"id" db:"task_id"`
	UserID      int64   `json:"-" db:"user_id"`
	Title       string  `json:"title" db:"title"`
	Status      int     `json:"-" db:"status"`
	Content     string  `json:"content,omitempty" db:"content"`
	CourseID    string  `json:"courseID,omitempty" db:"course_id"`
	Next        *string `json:"next,omitempty" db:"next"`
	CompletedAt int64   `json:"completedAt,omitempty" db:"completed_at"`
	Homework    *string `json:"homework,omitempty" db:"homework"` // Поле с выполненной домашней работой стажера
	IsFirst     bool    `json:"-" db:"is_first"`
	Cover       string  `json:"cover,omitempty" db:"cover"`
	Section     string  `json:"section,omitempty" db:"section"`
	Type        string  `json:"type" db:"type"`

	StatusPublic string `json:"status" db:"-"`
}
