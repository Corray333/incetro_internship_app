package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Corray333/internship_app/internal/types"
)

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
func findFirstTask(tasks []types.Task) *types.Task {
	for _, task := range tasks {
		if task.IsFirst {
			return &task
		}
	}
	return nil
}

// buildtypes.TaskMap строит карту задач по их types.TaskID для быстрого доступа
func buildTaskMap(tasks []types.Task) map[string]*types.Task {
	taskMap := make(map[string]*types.Task)
	for i := range tasks {
		taskMap[tasks[i].TaskID] = &tasks[i]
	}
	return taskMap
}

// topologicalSort выполняет топологическую сортировку задач
func TopologicalSort(tasks []types.Task) ([]types.Task, error) {
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

	firstTask := findFirstTask(tasks)
	if firstTask == nil {
		return nil, fmt.Errorf("first task not found")
	}

	if err := visit(firstTask); err != nil {
		return nil, err
	}

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
