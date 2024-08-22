package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"math/rand"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/labstack/echo/v4"
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

func gettasks() [][]string {
	r := make([][]string, 0)
	res, _ := db.FindAll(q.NewQuery(bucket))
	for i := range res {
		status := res[i].Get("status").(bool)
		r = append(r, []string{res[i].Get("id").(string), res[i].Get("name").(string), strconv.FormatBool(status)})
	}
	return r
}

func addtask(value string) {
	task := make(map[string]interface{})
	task["name"] = value
	task["status"] = false
	task["id"] = random()
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

	e.POST("/new", func(c echo.Context) error {
		task := c.FormValue("task")
		user := c.FormValue("user")
		password := c.FormValue("password")
		if user == cfg.username && password == cfg.password {
			if task != "" && len(strings.Split(task, "``")) == 1 {
				addtask(task)
				return c.String(http.StatusOK, "Success!\n")
			} else {
				return c.NoContent(http.StatusBadRequest)
			}
		}

		return c.NoContent(http.StatusForbidden)
	})

	e.POST("/list", func(c echo.Context) error {
		user := c.FormValue("user")
		password := c.FormValue("password")
		if user == cfg.username && password == cfg.password {
			var res string
			r := gettasks()
			for i := range r {
				res += fmt.Sprintf("%v``%v``%v\n", r[i][0], r[i][1], r[i][2])
			}
			return c.String(http.StatusOK, res)
		}
		return c.NoContent(http.StatusForbidden)
	})

	e.POST("/done", func(c echo.Context) error {
		user := c.FormValue("user")
		password := c.FormValue("password")
		id := c.FormValue("id")

		if user == cfg.username && password == cfg.password {
			update := make(map[string]interface{})
			update["status"] = true

			que := q.NewQuery(bucket).Where(q.Field("id").Eq(id))
			db.Update(que, update)

			return c.String(http.StatusOK, "Success!\n")
		}
		return c.NoContent(http.StatusForbidden)
	})

	e.POST("/delete", func(c echo.Context) error {
		user := c.FormValue("user")
		password := c.FormValue("password")
		id := c.FormValue("id")

		if user == cfg.username && password == cfg.password {
			que := q.NewQuery(bucket).Where(q.Field("id").Eq(id))
			db.Delete(que)

			return c.String(http.StatusOK, "Success!\n")
		}
		return c.NoContent(http.StatusForbidden)
	})

	e.POST("/reset", func(c echo.Context) error {
		user := c.FormValue("user")
		password := c.FormValue("password")
		id := c.FormValue("id")

		if user == cfg.username && password == cfg.password {
			update := map[string]interface{}{"status": false}
			que := q.NewQuery(bucket).Where(q.Field("id").Eq(id))
			db.Update(que, update)

			return c.String(http.StatusOK, "Success!\n")
		}
		return c.NoContent(http.StatusForbidden)
	})
	e.Logger.Fatal(e.Start(":" + cfg.port))
}
