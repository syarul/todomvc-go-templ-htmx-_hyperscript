package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type Todo struct {
	id      string
	title   string
	done    bool
	editing bool
}

type todos []Todo

type Filter struct {
	url      string
	name     string
	selected bool
}

var filters = []Filter{
	{url: "#/", name: "all", selected: true},
	{url: "#/active", name: "active", selected: false},
	{url: "#/completed", name: "completed", selected: false},
}

func main() {
	t := &todos{}

	http.Handle("/get-hash", http.HandlerFunc(getHash))
	http.Handle("/learn.json", http.HandlerFunc(learnHandler))

	http.Handle("/todo-filter", http.HandlerFunc(todoFilter))

	http.Handle("/", http.HandlerFunc(t.pageHandler))

	http.Handle("/add-todo", http.HandlerFunc(t.addTodoHandler))

	http.Handle("/completed", http.HandlerFunc(t.clearCompleted))

	http.Handle("/update-counts", http.HandlerFunc(t.updateCounts))

	http.Handle("/toggle-all", http.HandlerFunc(t.toggleAllHandler))

	http.Handle("/toggle-todo", http.HandlerFunc(t.toggleTodo))

	http.Handle("/edit-todo", http.HandlerFunc(t.editTodoHandler))

	http.Handle("/remove-todo", http.HandlerFunc(t.removeTodo))

	http.Handle("/update-todo", http.HandlerFunc(t.updateTodo))

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func learnHandler(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type header to indicate JSON
	w.Header().Set("Content-Type", "application/json")

	// Create an empty JSON object and write it to the response
	emptyJSON := map[string]interface{}{} // an empty JSON object
	json.NewEncoder(w).Encode(emptyJSON)
}

func todoFilter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	name := r.FormValue("name")

	// Loop through filters and update the selected field
	for i := range filters {
		if filters[i].name == name {
			filters[i].selected = true
		} else {
			filters[i].selected = false
		}
	}

	// Now 'filters' has been updated based on the selected filter name
	// You can use the updated filters as needed
	component := filter(filters)
	component.Render(r.Context(), w)
}

// countNotDone returns the count of todos that are not done
func countNotDone(todos []Todo) int {
	count := 0
	for _, todo := range todos {
		if !todo.done {
			count++
		}
	}
	return count
}

func defChecked(todos []Todo) bool {
	// Count the number of uncompleted tasks
	uncompletedCount := countNotDone(todos)

	// Determine the defaultChecked value
	defaultChecked := false
	if uncompletedCount == 0 && len(todos) > 0 {
		defaultChecked = true
	}
	return defaultChecked
}

func (t *todos) pageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := Page(*t, filters, defChecked(*t))
	component.Render(r.Context(), w)
}

func (t *todos) addTodoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	title := r.FormValue("title")

	randId := uuid.New().String()

	newTodo := Todo{
		id:      randId,
		title:   title,
		done:    false,
		editing: false,
	}

	*t = append(*t, newTodo)

	component := todoItem(newTodo)
	component.Render(r.Context(), w)
}

func getHash(w http.ResponseWriter, r *http.Request) {
	component := filter(filters)
	component.Render(r.Context(), w)
}

// hasIncompleteTask checks if there is any incomplete task in the Todos slice
func hasIncompleteTask(todos []Todo) bool {
	for _, todo := range todos {
		if todo.done {
			return true
		}
	}
	return false
}

func (t *todos) clearCompleted(w http.ResponseWriter, r *http.Request) {
	// Determine render "none" or "block" based on incomplete tasks
	displayStyle := "none"
	if hasIncompleteTask(*t) {
		displayStyle = "block"
	}

	// Write the string directly to the response
	w.Write([]byte(displayStyle))
}

func (t *todos) updateCounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	uncompletedCount := countNotDone(*t)
	plural := ""
	if uncompletedCount != 1 {
		plural = "s"
	}
	countString := fmt.Sprintf("<strong>%d</strong>", uncompletedCount)
	responseString := fmt.Sprintf("item%s left", plural)

	w.Write([]byte(countString + " " + responseString))
}

func (t *todos) toggleAllHandler(w http.ResponseWriter, r *http.Request) {
	// Count the number of uncompleted tasks
	checked := defChecked(*t)

	// Render the template or send the value to the client as needed
	w.Write([]byte(strconv.FormatBool(checked)))
}

func (t *todos) toggleTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	id := r.FormValue("id")
	done := r.FormValue("done")

	index := -1
	for i, todo := range *t {
		if todo.id == id {
			index = i
			break
		}
	}

	// If the todo is found, update its done status
	if index != -1 {
		if done == "true" {
			(*t)[index].done = false
		} else {
			(*t)[index].done = true
		}
		// render the new todo
		component := todoItem((*t)[index])
		component.Render(r.Context(), w)
	}
}

func (t *todos) editTodoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	id := r.FormValue("id")
	title := r.FormValue("title")

	index := -1
	for i, todo := range *t {
		if todo.id == id {
			index = i
			break
		}
	}

	if index != -1 {

		(*t)[index].title = title
		(*t)[index].editing = true
		// the trick is to only target the input element,
		// since there's bunch _hyperscript scope events happening here
		// we don't want to swap and loose the selectors.
		// Could also move it to the parentNode
		component := editTodo((*t)[index])
		component.Render(r.Context(), w)
	}
}

func (t *todos) removeTodo(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	index := -1
	for i, todo := range *t {
		if todo.id == id {
			index = i
			break
		}
	}

	if index != -1 {
		*t = append((*t)[:index], (*t)[index+1:]...)
	}

	w.Write([]byte(""))
}

func (t *todos) updateTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	id := r.FormValue("id")
	title := r.FormValue("title")

	index := -1
	for i, todo := range *t {
		if todo.id == id {
			index = i
			break
		}
	}

	// If the todo is found, update its title, editing
	if index != -1 {
		(*t)[index].title = title
		(*t)[index].editing = false

		// generate new todo templ
		component := todoItem((*t)[index])
		component.Render(r.Context(), w)
	}
}
