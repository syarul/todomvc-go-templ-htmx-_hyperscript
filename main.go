package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
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

// enum
type Action int

// group action related constants in one type
const (
	Create Action = 0
	Toggle Action = 1
	Edit   Action = 2
	Update Action = 3
	Delete Action = 4
)

var filters = []Filter{
	{url: "#/", name: "all", selected: true},
	{url: "#/active", name: "active", selected: false},
	{url: "#/completed", name: "completed", selected: false},
}

func (t *todos) crudOps(action Action, todo Todo) Todo {
	index := -1
	if action != Create {
		for i, r := range *t {
			if r.id == todo.id {
				index = i
				break
			}
		}
	}
	switch action {
	case Create:
		*t = append(*t, todo)
		return todo
	case Toggle:
		(*t)[index].done = todo.done
	case Update:
		(*t)[index].title = todo.title
		(*t)[index].editing = false
	case Delete:
		*t = append((*t)[:index], (*t)[index+1:]...)
	default:
		// Edit should do nothing only return todo from store
	}
	if index != -1 && action != Delete {
		return (*t)[index]
	}
	return Todo{}
}

// main is the entry point of the application.
func main() {
	t := &todos{}

	// Register the routes.
	http.Handle("/get-hash", http.HandlerFunc(getHash))
	http.Handle("/learn.json", http.HandlerFunc(learnHandler))

	http.Handle("/update-counts", http.HandlerFunc(t.updateCounts))
	http.Handle("/toggle-all", http.HandlerFunc(t.toggleAllHandler))
	http.Handle("/completed", http.HandlerFunc(t.clearCompleted))

	http.Handle("/", http.HandlerFunc(t.pageHandler))

	http.Handle("/add-todo", http.HandlerFunc(t.addTodoHandler))
	http.Handle("/toggle-todo", http.HandlerFunc(t.toggleTodo))
	http.Handle("/edit-todo", http.HandlerFunc(t.editTodoHandler))
	http.Handle("/update-todo", http.HandlerFunc(t.updateTodo))
	http.Handle("/remove-todo", http.HandlerFunc(t.removeTodo))

	// Start the server.
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
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

// hasIncompleteTask checks if there is any incomplete task in the Todos slice
func hasIncompleteTask(todos []Todo) bool {
	for _, todo := range todos {
		if todo.done {
			return true
		}
	}
	return false
}

// templRenderer sets the common headers and renders the given component.
func templRenderer(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component.Render(r.Context(), w)
}

// byteRenderer writes the given value as a response with the Content-Type header set to text/html; charset=utf-8.
func byteRenderer[V string](w http.ResponseWriter, r *http.Request, value V) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(value))
}

func learnHandler(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type header to indicate JSON
	w.Header().Set("Content-Type", "application/json")

	// Create an empty JSON object and write it to the response
	emptyJSON := map[string]interface{}{} // an empty JSON object
	json.NewEncoder(w).Encode(emptyJSON)
}

// getHash handles the GET request for the #/:name route.
// It updates the selected field of each filter based on the name query parameter.
func getHash(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	if name != "" {
		// Loop through filters and update the selected field
		for i := range filters {
			if filters[i].name == name {
				filters[i].selected = true
			} else {
				filters[i].selected = false
			}
		}
	}

	// Render the filter component with the updated filters
	templRenderer(w, r, filter(filters))
}

func (t *todos) pageHandler(w http.ResponseWriter, r *http.Request) {
	templRenderer(w, r, Page(*t, filters, defChecked(*t)))
}

func (t *todos) addTodoHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")

	// ignore adding if title is empty
	if len(title) == 0 {
		byteRenderer(w, r, "")
		return
	}

	id := uuid.New().String()

	todo := t.crudOps(Create, Todo{id, title, false, false})

	templRenderer(w, r, todoItem(todo))
}

func (t *todos) clearCompleted(w http.ResponseWriter, r *http.Request) {
	// Determine render "none" or "block" based on incomplete tasks
	displayStyle := "none"
	if hasIncompleteTask(*t) {
		displayStyle = "block"
	}

	// Write the string directly to the response
	byteRenderer(w, r, displayStyle)
}

func (t *todos) updateCounts(w http.ResponseWriter, r *http.Request) {
	uncompletedCount := countNotDone(*t)
	plural := ""
	if uncompletedCount != 1 {
		plural = "s"
	}

	byteRenderer(w, r, fmt.Sprintf("<strong>%d</strong> item%s left", uncompletedCount, plural))
}

func (t *todos) toggleAllHandler(w http.ResponseWriter, r *http.Request) {
	// Count the number of uncompleted tasks
	checked := defChecked(*t)

	// Render the template or send the value to the client as needed
	byteRenderer(w, r, strconv.FormatBool(checked))
}

func (t *todos) toggleTodo(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	done, err := strconv.ParseBool(r.FormValue("done"))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	todo := t.crudOps(Toggle, Todo{id, "", !done, false})

	templRenderer(w, r, todoItem(todo))
}

func (t *todos) editTodoHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	// the trick is to only target the input element,
	// since there's bunch _hyperscript scope events happening here
	// we don't want to swap and loose the selectors.
	// we also don't want to do any crud operations
	// since editing only client side changes
	todo := t.crudOps(Edit, Todo{id, "", false, false})

	templRenderer(w, r, editTodo(Todo{id, todo.title, todo.done, true}))
}

func (t *todos) updateTodo(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	title := r.FormValue("title")

	todo := t.crudOps(Update, Todo{id, title, false, false})

	templRenderer(w, r, todoItem(todo))
}

func (t *todos) removeTodo(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	t.crudOps(Delete, Todo{id, "", false, false})

	byteRenderer(w, r, "")
}
