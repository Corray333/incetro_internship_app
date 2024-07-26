package storage

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/Corray333/internship_app/internal/types"
	"github.com/Corray333/internship_app/pkg/server/auth"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sqlx.DB
	// courses map[string]string
}

func New() *Storage {
	db, err := sqlx.Open("postgres", os.Getenv("DB_CONN_STR"))
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return &Storage{
		db: db,
	}
}

func (s *Storage) CreateUser(user *types.User) error {

	fmt.Println()
	fmt.Printf("User creating: %+v\n", *user)
	fmt.Println()
	_, err := s.db.Exec(`
        INSERT INTO users (
            profile_id, 
            user_id, 
            person_id, 
            verified, 
            email, 
            course, 
            username, 
            phone, 
            fio, 
            avatar, 
            state, 
            fails
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `,
		user.ProfileID,
		user.UserID,
		user.PersonID,
		user.Verified,
		user.Email,
		user.Course,
		user.Username,
		user.Phone,
		user.FIO,
		user.Avatar,
		user.State,
		user.Fails)
	return err
}

func (s *Storage) VerifyUser(chat_id int64) error {
	_, err := s.db.Exec("UPDATE users SET verified = true WHERE user_id = $1", chat_id)
	return err
}

func (s *Storage) GetAllUsers() ([]types.User, error) {
	users := []types.User{}

	if err := s.db.Select(&users, "SELECT * FROM users"); err != nil {
		return nil, err
	}
	return users, nil
}

var courses = map[string]string{
	"iOS":     "1fd51f1d812049acb6bdcbfd681d472a",
	"Android": "9078876e1c7148c595d2b28f5e19faae",
	"Flutter": "9ea7bf749e47412bb7c99e99052dec8b",
	"Python":  "7323ca4bfd844d40bdbac557bd578bb0",
}

func (s *Storage) GetUsersOnCourse(course_id string) ([]types.User, error) {
	users := []types.User{}
	if course_id == "all" {
		if err := s.db.Select(&users, "SELECT * FROM users"); err != nil {
			return nil, err
		}
		return users, nil
	}
	course_id = courses[course_id]

	if err := s.db.Select(&users, "SELECT * FROM users WHERE course = $1", course_id); err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Storage) GetUserByID(user_id int64) (*types.User, error) {
	user := types.User{}
	err := s.db.QueryRowx(`
        SELECT 
            profile_id, 
            user_id, 
            person_id, 
            verified, 
            email, 
            course, 
            username, 
            phone, 
            fio, 
            avatar, 
            state, 
            fails 
        FROM 
            users 
        WHERE 
            user_id = $1
    `, user_id).StructScan(
		&user,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Storage) GetCourse(course_id string) (*types.Course, error) {
	course := &types.Course{}
	err := s.db.Get(course, "SELECT name, course_id, curator_id, short_name, invite_link, group_id FROM courses WHERE course_id = $1", course_id)
	return course, err
}

func (s *Storage) GetLastSynced() (int64, error) {
	var res int64
	err := s.db.QueryRow("SELECT last_synced FROM system WHERE id = 1").Scan(&res)
	return res, err
}

func (s *Storage) SetLastSynced(last_synced int64) error {
	_, err := s.db.Exec("UPDATE system SET last_synced = $1 WHERE id = 1", last_synced)
	return err
}

func (s *Storage) SetTask(notionID string, title string, content string, course_id string, next string, cover string, is_first bool, section string, task_type string) error {
	var nextTask *string
	if next == "" {
		nextTask = nil
	} else {
		nextTask = &next
	}
	_, err := s.db.Exec("INSERT INTO tasks (task_id, title, content, course_id, next, cover, is_first, section, type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (task_id) DO UPDATE SET title = $2, content = $3, course_id = $4, next = $5, cover = $6, is_first=$7, section=$8, type=$9", notionID, title, content, course_id, nextTask, cover, is_first, section, task_type)
	return err
}

func (s *Storage) SetCourse(course *types.Course) error {
	_, err := s.db.Exec("INSERT INTO courses (name, course_id, curator_id, short_name, invite_link, group_id) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (course_id) DO UPDATE SET name = $1, curator_id = $3, short_name=$4, invite_link=$5, group_id=$6", course.Name, course.NotionID, course.CuratorID, course.ShortName, course.Invite, course.GroupID)
	return err
}

func (s *Storage) GiveTasks(user_id int64, course_id string) error {
	_, err := s.db.Exec(`INSERT INTO progress (user_id, task_id, status)
	SELECT 
		u.user_id, 
		t.task_id, 
		1
	FROM 
		users u
	JOIN 
		tasks t ON u.course = t.course_id
	WHERE 
		u.user_id = $1 and t.course_id = $2`, user_id, course_id)

	if err != nil {
		return err
	}

	_, err = s.db.Exec("UPDATE progress SET status = 2 WHERE user_id = $1 AND progress.task_id = (SELECT task_id FROM tasks WHERE is_first = true AND course_id = (SELECT course FROM users WHERE user_id = $1))", user_id)

	return err
}

func (s *Storage) GetTasks(user_id int64) ([]types.Task, error) {
	tasks := []types.Task{}
	err := s.db.Select(&tasks, "SELECT task_id, title, status, type, is_first, next, section FROM progress NATURAL JOIN tasks WHERE progress.user_id = $1", user_id)
	for i := 0; i < len(tasks); i++ {
		tasks[i].StatusPublic = types.TaskStatuses[tasks[i].Status]
	}
	return tasks, err
}

func (s *Storage) GetTask(user_id int64, task_id string) (*types.Task, error) {
	task := types.Task{}
	err := s.db.Get(&task, "SELECT task_id, user_id, title, status, type, is_first, next, content, homework FROM progress NATURAL JOIN tasks WHERE progress.user_id = $1 AND task_id = $2", user_id, task_id)
	task.StatusPublic = types.TaskStatuses[task.Status]
	return &task, err

}

func (s *Storage) UpdateUser(user *types.User) error {
	_, err := s.db.Exec(`
        UPDATE users 
        SET 
            verified = $1, 
            email = $2, 
            course = $3, 
            person_id = $4, 
            username = $5, 
            phone = $6, 
            fio = $7, 
            avatar = $8, 
            state = $9, 
            fails = $10 
        WHERE 
            user_id = $11
    `,
		user.Verified,
		user.Email,
		user.Course,
		user.PersonID,
		user.Username,
		user.Phone,
		user.FIO,
		user.Avatar,
		user.State,
		user.Fails,
		user.UserID)
	return err
}

func (s *Storage) SaveHomework(userID int64, taskID string, homework string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec("UPDATE progress SET homework = $1, status=3 WHERE user_id = $2 AND task_id = $3", homework, userID, taskID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("wrong number of tasks got homework: " + strconv.Itoa(int(rows)))
	}

	return tx.Commit()
}

func (s *Storage) GetCuratorOfUser(user_id int64) (int64, error) {
	var res int64
	err := s.db.QueryRowx("SELECT curator_id FROM courses WHERE course_id = (SELECT course FROM users WHERE user_id = $1)", user_id).Scan(&res)
	return res, err
}

func (s *Storage) SetRefresh(uid int64, refresh string) error {
	_, err := s.db.Exec("INSERT INTO user_token VALUES ($1, $2)", uid, refresh)
	return err
}

func (s *Storage) RefreshToken(id int64, oldRefresh string) (string, string, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return "", "", err
	}

	rows := s.db.QueryRow(`
		SELECT token FROM user_token WHERE user_id = $1 AND token = $2;
	`, id, oldRefresh)

	var refresh string
	if err := rows.Scan(&refresh); err != nil {
		return "", "", err
	}
	if refresh != oldRefresh {
		fmt.Println()
		fmt.Println(refresh, " --- ", oldRefresh)
		fmt.Println()
		return "", "", fmt.Errorf("invalid refresh token")
	}

	newRefresh, err := auth.CreateToken(id, auth.RefreshTokenLifeTime)
	if err != nil {
		return "", "", err
	}

	newAccess, err := auth.CreateToken(id, auth.AccessTokenLifeTime)
	if err != nil {
		return "", "", err
	}

	_, err = tx.Queryx(`
		UPDATE user_token SET token = $1 WHERE user_id = $2 AND token = $3;
	`, newRefresh, id, oldRefresh)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(); err != nil {
		return "", "", err
	}

	return newAccess, newRefresh, nil
}

func (s *Storage) TaskDone(uid int64, task *types.Task) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
		UPDATE progress SET status=4, completed_at=$1 WHERE user_id=$2 AND task_id=$3
	`, task.CompletedAt, uid, task.TaskID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE progress SET status = 2 WHERE task_id = $1 and user_id = $2`, task.Next, uid)

	if err != nil {
		return err
	}

	return tx.Commit()

}

func (s *Storage) IsCurator(uid int64) (bool, error) {
	var res bool
	err := s.db.QueryRowx("SELECT EXISTS (SELECT 1 FROM courses WHERE curator_id = $1)", uid).Scan(&res)
	return res, err
}

func (s *Storage) RejectHomework(uid int64, taskID string) error {
	_, err := s.db.Exec("UPDATE progress SET status = 5 WHERE task_id = $1 and user_id = $2", taskID, uid)
	return err
}

func (s *Storage) SetUpdateData(updateID int, data string) error {
	_, err := s.db.Exec("INSERT INTO update_data VALUES($1,$2)", updateID, data)
	return err
}

func (s *Storage) GetUpdateData(updateID int) (string, error) {
	var data string
	err := s.db.QueryRowx("SELECT data FROM update_data WHERE update_id = $1", updateID).Scan(&data)
	return data, err
}
