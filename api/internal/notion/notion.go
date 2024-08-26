package notion

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Corray333/internship_app/internal/types"
	"github.com/Corray333/internship_app/internal/utils"
	"github.com/Corray333/internship_app/pkg/notion"
	"github.com/spf13/viper"
)

const TIME_LAYOUT = "2006-01-02T15:04:05.000-07:00"
const TIME_LAYOUT_IN = "2006-01-02T15:04:05.999Z07:00"

type Storage interface {
	GetLastSynced() (int64, error)
	SetTask(notionID string, title string, content string, course_id string, next string, cover string, is_first bool, section string, task_type string) error
	SetLastSynced(last_synced int64) error
	SetCourse(course *types.Course) error
}

func CreateUser(chatID int64, user *types.User) (string, error) {
	req := map[string]interface{}{
		"Имя": map[string]interface{}{
			"title": []map[string]interface{}{
				{
					"text": map[string]string{
						"content": user.FIO,
					},
				},
			},
		},
		"Telegram": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": "@" + user.Username,
						"link": map[string]interface{}{
							"url": "https://t.me/" + user.Username,
						},
					},
				},
			},
		},
		"Курс": map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": user.Course,
				},
			},
		},
		"Почта": map[string]interface{}{
			"type":  "email",
			"email": user.Email,
		},
		"Телефон": map[string]interface{}{
			"type":         "phone_number",
			"phone_number": user.Phone,
		},
	}

	resp, err := notion.CreatePage(viper.GetString("interns_table"), req, user.Avatar)
	if err != nil {
		return "", err
	}
	userId := struct {
		ID string `json:"id"`
	}{}
	if err := json.Unmarshal(resp, &userId); err != nil {
		return "", err
	}

	return userId.ID, nil
}

type Tasks struct {
	Results    []Task `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

type Task struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_time"`
	UpdatedAt string `json:"last_edited_time"`
	Cover     struct {
		External struct {
			URL string `json:"url"`
		} `json:"external"`
	} `json:"cover"`
	Properties struct {
		TP struct {
			People []struct {
				ID string `json:"id"`
			} `json:"people"`
		} `json:"TP"`
		Skill struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Навык"`
		TypeField struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Тип"`
		NextStep struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Следующий шаг"`
		PrevStep struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Предыдущий шаг"`
		Direction struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Направление"`
		Section struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Группа"`
	} `json:"properties"`
}

type Profile struct {
	Results []struct {
		Properties struct {
			Profile struct {
				People []struct {
					ID string `json:"id"`
				} `json:"people"`
			} `json:"Профиль"`
		} `json:"properties"`
	} `json:"results"`
}

type Courses struct {
	Results []Course `json:"results"`
}

type Course struct {
	ID         string `json:"id"`
	Properties struct {
		Invitation struct {
			RichText []struct {
				PlainText string `json:"plain_text"`
			} `json:"rich_text"`
		} `json:"Приглашение"`
		Admin struct {
			Number int64 `json:"number"`
		} `json:"Админ"`
		Direction struct {
			RichText []struct {
				PlainText string `json:"plain_text"`
			} `json:"rich_text"`
		} `json:"Направление"`
		Group struct {
			Number int64 `json:"number"`
		} `json:"Группа"`
		Launched struct {
			Checkbox bool `json:"checkbox"`
		} `json:"Запущен"`
		Title struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Название"`
	} `json:"properties"`
}

func GetCourses() (map[string]types.Course, error) {
	resp, err := notion.SearchPages(viper.GetString("courses_table"), map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "Запущен",
			"checkbox": map[string]bool{
				"equals": true,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	courses := Courses{}
	if err := json.Unmarshal(resp, &courses); err != nil {
		return nil, err
	}
	result := map[string]types.Course{}

	for _, course := range courses.Results {
		result[course.ID] = types.Course{
			Name:      course.Properties.Title.Title[0].PlainText,
			NotionID:  course.ID,
			CuratorID: course.Properties.Admin.Number,
			ShortName: course.Properties.Direction.RichText[0].PlainText,
			GroupID:   course.Properties.Group.Number,
			Invite:    course.Properties.Invitation.RichText[0].PlainText,
		}
	}
	return result, nil
}

func GetTasks(lastSynced int64) ([]Task, error) {
	fmt.Println(time.Unix(lastSynced, 0).Format(TIME_LAYOUT_IN))

	var allTasks []Task
	var nextCursor string
	hasMore := true

	for hasMore {
		// Prepare the request payload
		payload := map[string]interface{}{
			"filter": map[string]interface{}{
				"and": []map[string]interface{}{
					{
						"property": "Курс запущен",
						"rollup": map[string]interface{}{
							"any": map[string]interface{}{
								"checkbox": map[string]bool{
									"equals": true,
								},
							},
						},
					},
					{
						"timestamp": "last_edited_time",
						"last_edited_time": map[string]interface{}{
							"after": time.Unix(lastSynced, 0).Format(TIME_LAYOUT_IN),
						},
					},
				},
			},
			"sorts": []map[string]interface{}{
				{
					"timestamp": "last_edited_time",
					"direction": "ascending",
				},
			},
		}

		// Add the next cursor if it's not the first request
		if nextCursor != "" {
			payload["start_cursor"] = nextCursor
		}

		// Fetch the pages from Notion
		resp, err := notion.SearchPages(viper.GetString("tasks_table"), payload)
		if err != nil {
			return nil, err
		}

		// Unmarshal the response
		tasks := &Tasks{}
		if err := json.Unmarshal(resp, tasks); err != nil {
			return nil, err
		}

		// Append the results to the allTasks slice
		allTasks = append(allTasks, tasks.Results...)

		// Update hasMore and nextCursor for the next iteration
		hasMore = tasks.HasMore
		nextCursor = tasks.NextCursor
	}

	return allTasks, nil
}

func LoadTasks(store Storage) error {
	lastSynced, err := store.GetLastSynced()
	if err != nil {
		return err
	}
	tasks, err := GetTasks(lastSynced)
	if err != nil {
		return err
	}
	fmt.Println(len(tasks))
	for _, task := range tasks {
		if len(task.Properties.Direction.Relation) == 0 || (len(task.Properties.NextStep.Relation) == 0 && len(task.Properties.PrevStep.Relation) == 0) {
			continue
		}
		content, err := GetMarkdown(task.ID)
		if err != nil {
			return err
		}

		title := ""
		for _, part := range task.Properties.Skill.Title {
			title += part.PlainText
		}
		if len(task.Properties.NextStep.Relation) == 0 {
			task.Properties.NextStep.Relation = make([]struct {
				ID string "json:\"id\""
			}, 1)
			task.Properties.NextStep.Relation[0].ID = ""
		}

		content, err = utils.ProcessMarkdown(content)
		if err != nil {
			return err
		}

		if err := store.SetTask(task.ID, title, content, task.Properties.Direction.Relation[0].ID, task.Properties.NextStep.Relation[0].ID, task.Cover.External.URL, len(task.Properties.PrevStep.Relation) == 0, task.Properties.Section.Select.Name, task.Properties.TypeField.Select.Name); err != nil {
			return fmt.Errorf("error while saving task %+v in store: %w", task, err)
		}
		last_synced, err := time.Parse(TIME_LAYOUT_IN, task.UpdatedAt)
		if err != nil {
			return fmt.Errorf("time %s has wrong created time format: %w", task.ID, err)
		}
		lastSynced = last_synced.Unix()
		if err := store.SetLastSynced(lastSynced); err != nil {
			return err
		}
	}

	return nil
}

func SyncCourses(store Storage) error {
	courses, err := GetCourses()
	if err != nil {
		return err
	}
	for _, v := range courses {
		if err := store.SetCourse(&v); err != nil {
			return err
		}
	}
	return nil
}

func Sync(store Storage) error {
	if err := SyncCourses(store); err != nil {
		return err
	}
	if err := LoadTasks(store); err != nil {
		return err
	}

	return nil
}

type Page struct {
	MD string `json:"md"`
}

func GetMarkdown(pageID string) (string, error) {
	url := viper.GetString("node_url") + "get-md?pageId=" + pageID
	fmt.Println(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("notion error %s while getting markdown of page: %s", string(body), pageID)
	}

	page := Page{}

	if err := json.Unmarshal(body, &page); err != nil {
		return "", err
	}

	fmt.Println()
	fmt.Println(page.MD)
	fmt.Println()

	return page.MD, nil
}
