package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/labstack/echo/v4"
)

const bucket = "tasks"

var db, _ = bolt.Open("db", 0600, nil)

func setupdb() {
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(bucket))
		if err == nil {
			b.Put([]byte("l"), []byte("0"))
		}
		return nil
	})
}

func put(key, value string) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		b.Put([]byte(key), []byte(value))
		return nil
	})
}

func get(key string) []byte {
	r := make([]byte, 0)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		r = b.Get([]byte(key))
		return nil
	})

	return r
}

func getall() []string {
	r := make([]string, 0)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))

		b.ForEach(func(k, v []byte) error {
			r = append(r, string(v))
			return nil
		})

		return nil
	})

	return r
}

func gettasks() []string {
	r := make([]string, 0)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))

		b.ForEach(func(k, v []byte) error {
			if string(k) != "l" {
				r = append(r, string(v))
			}
			return nil
		})

		return nil
	})

	return r
}

func addtask(value string) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		tkey, _ := strconv.Atoi(string(b.Get([]byte("l"))))
		tkey++
		key := strconv.Itoa(tkey)
		b.Put([]byte(key), []byte(value+"``0"))
		b.Put([]byte("l"), []byte(key))
		return nil
	})
}

func main() {
	defer db.Close()
	setupdb()

	data, err := os.ReadFile("auth")
	if err != nil {
		log.Fatalf("Error while loading auth: %v", err)
	}
	auth := strings.Split(string(data), ":")

	e := echo.New()

	e.POST("/new", func(c echo.Context) error {
		task := c.FormValue("task")
		user := c.FormValue("user")
		password := c.FormValue("password")
		if user == strings.Trim(auth[0], "\n") && password == strings.Trim(auth[1], "\n") {
			addtask(task)
			return c.String(http.StatusOK, "Success!\n")
		}
		return c.String(http.StatusOK, "Failed!\n")
	})

	e.POST("/list", func(c echo.Context) error {
		user := c.FormValue("user")
		password := c.FormValue("password")
		res := "Tasks:\n"
		r := make([]string, 0)
		if user == strings.Trim(auth[0], "\n") && password == strings.Trim(auth[1], "\n") {
			r = gettasks()
			for i := range r {
				t := strings.Split(r[i], "``")
				var status string
				if t[1] == "0" {
					status = "Not finished"
				} else {
					status = "Finished"
				}
				res += fmt.Sprintf("%v ----- %v\n", t[0], status)
			}
			return c.String(http.StatusOK, res)
		}
		return c.String(http.StatusOK, "Failed!\n")
	})
	e.Logger.Fatal(e.Start(":3333"))
}
