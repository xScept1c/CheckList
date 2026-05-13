package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

var tasks = []Task{}
var nextID = 1

func main() {
	loadTasks()

	go func() {
		http.HandleFunc("/", indexHandler)
		http.HandleFunc("/api/tasks", tasksHandler)
		http.HandleFunc("/api/add", addHandler)
		http.HandleFunc("/api/delete/", deleteHandler)
		http.HandleFunc("/api/toggle/", toggleHandler)
		http.HandleFunc("/api/update/", updateHandler)

		fmt.Println("✅ Сервер запущен на http://localhost:8080")
		http.ListenAndServe(":8080", nil)
	}()

	time.Sleep(1 * time.Second)

	openBrowser("http://localhost:8080")

	fmt.Println("📁 Данные сохраняются в tasks.json")
	fmt.Print("❌ Нажмите Ctrl+C для выхода\n")

	select {}
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin": // macOS
		err = exec.Command("open", url).Start()
	default: // Linux и другие
		err = exec.Command("xdg-open", url).Start()
	}

	if err != nil {
		fmt.Printf("❌ Не удалось открыть браузер автоматически\n")
		fmt.Printf("🌐 Пожалуйста, откройте вручную: %s\n", url)
	} else {
		fmt.Printf("🌐 Браузер открыт: %s\n", url)
	}
}

func loadTasks() {
	data, _ := ioutil.ReadFile("tasks.json")
	json.Unmarshal(data, &tasks)
	for _, t := range tasks {
		if t.ID >= nextID {
			nextID = t.ID + 1
		}
	}
}

func saveTasks() {
	data, _ := json.MarshalIndent(tasks, "", "  ")
	ioutil.WriteFile("tasks.json", data, 0644)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(tasks)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	var t Task
	json.NewDecoder(r.Body).Decode(&t)
	t.ID = nextID
	t.Status = "pending"
	nextID++
	tasks = append(tasks, t)
	saveTasks()
	json.NewEncoder(w).Encode(t)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/delete/"))
	for i, t := range tasks {
		if t.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			break
		}
	}
	saveTasks()
	w.WriteHeader(200)
}

func toggleHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/toggle/"))
	for i, t := range tasks {
		if t.ID == id {
			if t.Status == "pending" {
				tasks[i].Status = "completed"
			} else {
				tasks[i].Status = "pending"
			}
			break
		}
	}
	saveTasks()
	w.WriteHeader(200)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/update/"))
	var update Task
	json.NewDecoder(r.Body).Decode(&update)
	for i, t := range tasks {
		if t.ID == id {
			if update.Title != "" {
				tasks[i].Title = update.Title
			}
			if update.Description != "" {
				tasks[i].Description = update.Description
			}
			if update.Status != "" {
				tasks[i].Status = update.Status
			}
			break
		}
	}
	saveTasks()
	w.WriteHeader(200)
}
