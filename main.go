package main

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"

	"math/rand"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	c "github.com/ostafen/clover/v2"
	d "github.com/ostafen/clover/v2/document"
	q "github.com/ostafen/clover/v2/query"
)

const bucket = "tasks"

var db, _ = c.Open("")

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var conf = koanf.New(".")

type config struct {
	port     string
	username string
	password string
}

type Template struct {
	Templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}

func NewTemplateRenderer(e *echo.Echo, paths ...string) {
	tmpl := &template.Template{}
	for i := range paths {
		template.Must(tmpl.ParseGlob(paths[i]))
	}
	t := newTemplate(tmpl)
	e.Renderer = t
}

func newTemplate(templates *template.Template) echo.Renderer {
	return &Template{
		Templates: templates,
	}
}

var cfg = load_config()

func load_config() config {
	config_path := "/etc/doit/config.yaml"
	if err := conf.Load(file.Provider(config_path), yaml.Parser()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg := config{}
	cfg.port = conf.String("port")
	cfg.username = conf.String("username")
	cfg.password = conf.String("password")

	return cfg
}

func random() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func setupdb() {
	truthy, _ := db.HasCollection(bucket)
	if !truthy {
		db.CreateCollection(bucket)
	}
}

func checkids(id string) bool {
	res, _ := db.FindAll(q.NewQuery(bucket).Where(q.Field("id").Eq(id)))
	r := true
	if len(res) == 0 {
		r = false
	}

	return r
}

type Task struct {
	Id     string
	Tag    string
	Name   string
	Status string
}

func gettasks() [][]string {
	r := make([][]string, 0)
	res, _ := db.FindAll(q.NewQuery(bucket))
	for i := range res {
		status := res[i].Get("status").(bool)
		r = append(r, []string{res[i].Get("id").(string), res[i].Get("name").(string), strconv.FormatBool(status), res[i].Get("tag").(string)})
	}
	return r
}

func addtask(value, tag string) {
	task := make(map[string]interface{})
	task["name"] = value
	task["status"] = false
	task["id"] = random()
	task["tag"] = tag
	task["note"] = ""
	for checkids(task["id"].(string)) { // There is a little to no chance this will ever generate same key twice, but better be safe
		task["id"] = random()
	}
	doc := d.NewDocumentOf(task)
	db.InsertOne(bucket, doc)
}

func main() {
	defer db.Close()
	setupdb()

	e := echo.New()
	NewTemplateRenderer(e, "public/*.html")
	e.Static("/static", "static/")
	e.Use(middleware.BasicAuth(func(s1, s2 string, ctx echo.Context) (bool, error) {
		if subtle.ConstantTimeCompare([]byte(s1), []byte(cfg.username)) == 1 && subtle.ConstantTimeCompare([]byte(s2), []byte(cfg.password)) == 1 {
			return true, nil
		}
		return false, nil
	}))

	e.RouteNotFound("/*", func(c echo.Context) error {
		return c.String(http.StatusNotFound, "Are you stupid or something? Oh, you hate reading documentation. Just read it!")
	})

	e.GET("/", func(c echo.Context) error {
		r := make([]Task, 0)
		tasks := gettasks()
		for i := range tasks {
			r = append(r, Task{Id: tasks[i][0], Name: tasks[i][1], Status: tasks[i][2], Tag: tasks[i][3]})
		}
		res := map[string]interface{}{
			"items": r,
		}
		return c.Render(http.StatusOK, "index", res)
	})

	e.POST("/new", func(c echo.Context) error {
		task := c.FormValue("task")
		tag := c.FormValue("tag")
		if task != "" && len(strings.Split(task, "``")) == 1 {
			addtask(task, tag)
			return c.NoContent(http.StatusOK)
		} else {
			return c.NoContent(http.StatusBadRequest)
		}
	})

	e.POST("/list", func(c echo.Context) error {
		var res string
		r := gettasks()
		for i := range r {
			res += fmt.Sprintf("%v``%v``%v``%v\n", r[i][0], r[i][1], r[i][2], r[i][3])
		}
		return c.String(http.StatusOK, res)
	})

	e.POST("/done", func(c echo.Context) error {
		id := c.FormValue("id")

		update := make(map[string]interface{})
		update["status"] = true

		que := q.NewQuery(bucket).Where(q.Field("id").Eq(id))
		db.Update(que, update)

		return c.NoContent(http.StatusOK)
	})

	e.POST("/delete", func(c echo.Context) error {
		id := c.FormValue("id")

		que := q.NewQuery(bucket).Where(q.Field("id").Eq(id))
		db.Delete(que)

		return c.NoContent(http.StatusOK)
	})

	e.POST("/reset", func(c echo.Context) error {
		id := c.FormValue("id")

		update := map[string]interface{}{"status": false}
		que := q.NewQuery(bucket).Where(q.Field("id").Eq(id))
		db.Update(que, update)

		return c.NoContent(http.StatusOK)
	})

	e.POST("/rename", func(c echo.Context) error {
		id := c.FormValue("id")
		task := c.FormValue("task")
		update := map[string]interface{}{"name": task}
		que := q.NewQuery(bucket).Where(q.Field("id").Eq(id))
		db.Update(que, update)

		return c.NoContent(http.StatusOK)

	})

	e.POST("/getnote", func(c echo.Context) error {
		id := c.FormValue("id")
		que := q.NewQuery(bucket).Where(q.Field("id").Eq(id))
		d, err := db.FindFirst(que)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		return c.String(http.StatusOK, d.Get("note").(string))
	})

	e.POST("/newnote", func(c echo.Context) error {
		id := c.FormValue("id")
		note := c.FormValue("note")
		update := map[string]interface{}{"note": note}
		que := q.NewQuery(bucket).Where(q.Field("id").Eq(id))
		db.Update(que, update)

		return c.NoContent(http.StatusOK)
	})

	e.POST("/deletenote", func(c echo.Context) error {
		id := c.FormValue("id")
		update := map[string]interface{}{"note": ""}
		que := q.NewQuery(bucket).Where(q.Field("id").Eq(id))
		db.Update(que, update)

		return c.NoContent(http.StatusOK)
	})

	e.POST("/edittag", func(c echo.Context) error {
		id := c.FormValue("id")
		tag := c.FormValue("tag")
		update := map[string]interface{}{"tag": tag}
		que := q.NewQuery(bucket).Where(q.Field("id").Eq(id))
		db.Update(que, update)

		return c.NoContent(http.StatusOK)
	})
	e.Logger.Fatal(e.Start(":" + cfg.port))
}
