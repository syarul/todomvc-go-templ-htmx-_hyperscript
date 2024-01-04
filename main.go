package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/a-h/templ"
)

var idCounter uint64

type Todo struct {
	id      uint64
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
	{url: "#/", name: "All", selected: true},
	{url: "#/active", name: "Active", selected: false},
	{url: "#/completed", name: "Completed", selected: false},
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
		title := strings.Trim(todo.title, " ")
		if len(title) != 0 {
			(*t)[index].title = title
			(*t)[index].editing = false
		} else {
			// remove if title is empty
			*t = append((*t)[:index], (*t)[index+1:]...)
			return Todo{}
		}
	case Delete:
		*t = append((*t)[:index], (*t)[index+1:]...)
	default:
		// edit should do nothing only return todo from store
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
	http.Handle("/get-hash", http.HandlerFunc(t.getHash))
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

	http.Handle("/toggle-main", http.HandlerFunc(t.toggleMainHandler))
	http.Handle("/toggle-footer", http.HandlerFunc(t.toggleFooterHandler))

	// Specify the directory containing your static files
	dir := "./assets"

	// Use the http.FileServer to create a handler for serving static files
	fs := http.FileServer(http.Dir(dir))

	// Use the http.Handle to register the file server handler for a specific route
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// start the server.
	addr := os.Getenv("LISTEN_ADDRESS")
	if addr == "" {
		addr = "localhost:8888"
	}

	fmt.Printf("Listening on %s\n", addr)

	// Start the HTTP server
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
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
	// count the number of uncompleted tasks
	uncompletedCount := countNotDone(todos)

	// determine the defaultChecked value
	defaultChecked := false
	if uncompletedCount == 0 && len(todos) > 0 {
		defaultChecked = true
	}
	return defaultChecked
}

// has completeTask checks if there is any completed task in the Todos slice
func hasCompleteTask(todos []Todo) bool {
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
	// set the Content-Type header to indicate JSON
	w.Header().Set("Content-Type", "application/json")

	// create an empty JSON object and write it to the response
	emptyJSON := map[string]interface{}{} // an empty JSON object
	json.NewEncoder(w).Encode(emptyJSON)
}

// getHash handles the GET request for the #/:name route.
// it updates the selected field of each filter based on the name query parameter.
// on initial fetch when todos is empty it will send "empty string" which usually
// ignored by htmx, the reason here we want to make use templ rendering to behave
// efficiently by rendering only needed html, this use case usually available in
// modern client framework i.e., in React you can do
// <>{todo.length && <TodoList />}</>
// so you can do the same in HTMX too!
func (t *todos) getHash(w http.ResponseWriter, r *http.Request) {
	if len(*t) == 0 {
		byteRenderer(w, r, "")
		return
	}

	hash := r.FormValue("hash")
	name := r.FormValue("name")

	if len(name) == 0 {
		if len(hash) != 0 {
			name = hash
		} else {
			name = "All"
		}
	}
	// loop through filters and update the selected field
	for i := range filters {
		if filters[i].name == name {
			filters[i].selected = true
		} else {
			filters[i].selected = false
		}
	}

	// render the filter component with the updated filters
	templRenderer(w, r, filter(filters))
}

func (t *todos) pageHandler(w http.ResponseWriter, r *http.Request) {
	// start with new todo data when refresh
	// *t = make([]Todo, 0)
	// idCounter = 0
	templRenderer(w, r, Page(*t, filters, defChecked(*t), hasCompleteTask(*t)))
}

// toggle section main
func (t *todos) toggleMainHandler(w http.ResponseWriter, r *http.Request) {
	templRenderer(w, r, toggleMain(*t, defChecked(*t)))
}

// toggle footer footer
func (t *todos) toggleFooterHandler(w http.ResponseWriter, r *http.Request) {
	templRenderer(w, r, footer(*t, filters, hasCompleteTask(*t)))
}

func (t *todos) addTodoHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.Trim(r.FormValue("title"), " ")
	// ignore adding if title is empty
	if len(title) == 0 {
		byteRenderer(w, r, "")
		return
	}

	idCounter++
	id := idCounter

	todo := t.crudOps(Create, Todo{id, title, false, false})

	if len(*t) == 1 {
		templRenderer(w, r, todoList(*t))
	} else {
		templRenderer(w, r, todoItem(todo))
	}

}

func (t *todos) clearCompleted(w http.ResponseWriter, r *http.Request) {
	// determine render "none" or "block" based on incomplete tasks
	hasCompleted := hasCompleteTask(*t)
	if hasCompleted {
		templRenderer(w, r, clearCompleted(hasCompleteTask(*t)))
	} else {
		byteRenderer(w, r, "")
	}
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
	// count the number of uncompleted tasks
	checked := defChecked(*t)

	// render the template or send the value to the client as needed
	byteRenderer(w, r, strconv.FormatBool(checked))
}

func (t *todos) toggleTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.FormValue("id"), 0, 32)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	done, err := strconv.ParseBool(r.FormValue("done"))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	todo := t.crudOps(Toggle, Todo{id, "", !done, false})

	templRenderer(w, r, todoItem(todo))
}

func (t *todos) editTodoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.FormValue("id"), 0, 32)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// the trick is to only target the input element,
	// since there's bunch _hyperscript scope events happening here
	// we don't want to swap and loose the selectors.
	// we also don't want to do any crud operations
	// since editing only client side changes
	todo := t.crudOps(Edit, Todo{id, "", false, false})

	templRenderer(w, r, editTodo(Todo{id, todo.title, todo.done, true}))
}

func (t *todos) updateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.FormValue("id"), 0, 32)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// esc, err := strconv.ParseBool(r.FormValue("esc"))
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	title := r.FormValue("title")

	// esc will not do update instead will send the original unmodified title
	// if esc {
	// 	todo := t.crudOps(Edit, Todo{id, "", false, false})
	// 	templRenderer(w, r, todoItem(todo))
	// } else {
	todo := t.crudOps(Update, Todo{id, title, false, false})
	if len(todo.title) == 0 {
		byteRenderer(w, r, "")
		return
	}
	templRenderer(w, r, todoItem(todo))
	// }
}

func (t *todos) removeTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.FormValue("id"), 0, 32)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	t.crudOps(Delete, Todo{id, "", false, false})

	byteRenderer(w, r, "")
}
