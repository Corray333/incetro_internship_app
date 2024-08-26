package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/Corray333/internship_app/internal/types"
)

func GroupTasks(tasks []types.Task) [][]types.Task {
	if len(tasks) == 0 {
		return [][]types.Task{}
	}
	sections := [][]types.Task{
		{
			tasks[0],
		},
	}
	currentSection := tasks[0].Section
	for _, task := range tasks[1:] {
		if task.Section == currentSection {
			sections[len(sections)-1] = append(sections[len(sections)-1], task)
		} else {
			sections = append(sections, []types.Task{task})
			currentSection = task.Section
		}
	}
	return sections
}

func DownloadImage(url string) error {
	// Create the images directory if it doesn't exist
	err := os.MkdirAll("../public/images", os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Get the response bytes from the URL
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download image: %v", err)
	}
	defer response.Body.Close()

	// Create the file
	filename := filepath.Base(url)
	filepath := filepath.Join("../public/images", filename)
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("failed to write image to file: %v", err)
	}

	fmt.Println("Image downloaded and saved to:", filepath)
	return nil
}

// findFirsttypes.Task находит первую задачу в слайсе задач
func findStartingTask(tasks []types.Task) *types.Task {
	// Создаем множество задач, которые являются значениями в поле Next у других задач
	nextTasks := make(map[string]bool)
	for _, task := range tasks {
		if task.Next != nil {
			nextTasks[*task.Next] = true
		}
	}

	// Ищем задачу, которая не указана в поле Next ни у одной из задач
	for _, task := range tasks {
		if !nextTasks[task.TaskID] {
			return &task
		}
	}

	return nil
}

// buildTaskMap строит карту задач по их types.TaskID для быстрого доступа
func buildTaskMap(tasks []types.Task) map[string]*types.Task {
	taskMap := make(map[string]*types.Task)
	for i := range tasks {
		taskMap[tasks[i].TaskID] = &tasks[i]
	}
	return taskMap
}

// TopologicalSort выполняет топологическую сортировку задач
func TopologicalSort(tasks []types.Task) ([]types.Task, error) {
	if len(tasks) == 0 {
		return []types.Task{}, nil
	}
	taskMap := buildTaskMap(tasks)
	var sortedTasks []types.Task
	visited := make(map[string]bool)

	var visit func(task *types.Task) error
	visit = func(task *types.Task) error {
		if visited[task.TaskID] {
			return nil
		}
		visited[task.TaskID] = true

		if task.Next != nil {
			nextTask, exists := taskMap[*task.Next]
			if exists {
				if err := visit(nextTask); err != nil {
					return err
				}
			}
		}

		sortedTasks = append(sortedTasks, *task)
		return nil
	}

	startingTask := findStartingTask(tasks)
	if startingTask == nil {
		return nil, fmt.Errorf("starting task not found")
	}

	if err := visit(startingTask); err != nil {
		return nil, err
	}

	// Разворачиваем список, так как задачи были добавлены в обратном порядке
	slices.Reverse(sortedTasks)
	return sortedTasks, nil
}

func EscapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

func ProcessMarkdown(markdown string) (string, error) {
	saveDir := "../public/files/"
	serverURL := os.Getenv("FILES_URL")
	// Регулярное выражение для поиска ссылок на файлы
	re := regexp.MustCompile(`\((https://prod-files-secure[^\s)]+)\)`)

	// Поиск всех совпадений
	matches := re.FindAllStringSubmatch(markdown, -1)

	// Пробегаем по всем найденным ссылкам
	for _, match := range matches {
		fileURL := match[1]
		if fileURL == "" {
			fileURL = match[3]
		}

		newFileName, err := SaveFileFromURL(fileURL, saveDir)
		if err != nil {
			return "", err
		}

		// Заменяем старую ссылку на новую
		newURL := serverURL + newFileName
		markdown = strings.Replace(markdown, fileURL, newURL, -1)
	}

	return markdown, nil
}

// SaveFileFromURL загружает файл по URL и сохраняет его с новым уникальным именем
func SaveFileFromURL(fileURL, saveDir string) (string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}

	// Получаем имя файла из URL
	fileName := filepath.Base(parsedURL.Path)

	// Генерируем случайную строку для уникального имени файла
	randomSuffix := generateRandomString(16)
	ext := filepath.Ext(fileName)
	name := fileName[:len(fileName)-len(ext)]
	newFileName := fmt.Sprintf("%s_%s%s", name, randomSuffix, ext)

	// Загружаем файл
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: %s", fileURL)
	}

	// Создаем файл в директории сохранения
	newFilePath := filepath.Join(saveDir, newFileName)
	newFile, err := os.Create(newFilePath)
	if err != nil {
		return "", err
	}
	defer newFile.Close()

	// Копируем содержимое загруженного файла в новый файл
	_, err = io.Copy(newFile, resp.Body)
	if err != nil {
		return "", err
	}

	return newFileName, nil
}

// generateRandomString генерирует случайную строку заданной длины
func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	for i := 0; i < n; i++ {
		bytes[i] = letters[int(bytes[i])%len(letters)]
	}
	return string(bytes)
}
