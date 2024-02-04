
     ooooo   ooooo ooooooooooooo ooo        ooooo ooooooo  ooooo 
     `888'   `888' 8'   888   `8 `88.       .888'  `8888    d8'  
      888     888       888       888b     d'888     Y888..8P    
      888ooooo888       888       8 Y88. .P  888      `8888'     
      888     888       888       8  `888'   888     .8PY888.    
      888     888       888       8    Y     888    d8'  `888b   
     o888o   o888o     o888o     o8o        o888o o888o  o88888o
    ===========================================================
            Build with GO, TEMPL, HTMX & _HYPERSCRIPT
[![Go](https://github.com/syarul/todomvc-go-templ-htmx-_hyperscript/actions/workflows/go.yml/badge.svg)](https://github.com/syarul/todomvc-go-templ-htmx-_hyperscript/actions/workflows/go.yml)

### E2E Testing


https://github.com/syarul/todomvc-go-templ-htmx-_hyperscript/assets/2774594/fdcba602-73f2-499b-a106-152569d37e80


Emulating the functionalities of modern frameworks which is base on React TodoMVC, [cypress test](https://github.com/syarul/todomvc-go-templ-htmx-_hyperscript/actions/runs/7412273948/job/20168687544) from https://github.com/cypress-io/cypress-example-todomvc. This demonstration serves to showcase that HTMX, when paired with _hyperscript, can replicate if not all the behaviors typically associated with most modern client framework with minimum needs to write javascript.

### Security
Check on [this link](https://templ.guide/security/) when using `templ` as HTML template engine.

### Usage
- install go if you don't have
- run `go mod tidy` to fetch all needed modules
- install templ `go install github.com/a-h/templ/cmd/templ@latest`
- run `templ generate`
- finally run `go run .`
- visit `http://localhost:8888`
- alternatively you can compile into executable with `go build .`
- if you need to run the e2e testing make sure to have `nodejs` installed
- run in the root folder `git clone https://github.com/cypress-io/cypress-example-todomvc`
- `cd cypress-example-todomvc`
- `npm install`
- if you need to see the test in browser run `npm run cypress:open`
- for headless test `npm run cypress:run`

### HTMX
Visit [https://github.com/rajasegar/awesome-htmx](https://github.com/rajasegar/awesome-htmx) to look for HTMX curated infos

###
Todo
- Perf test (consolidate with other langs rust, zig, odin, ocaml, etc+)
