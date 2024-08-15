package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/labstack/echo/v4"
)

const bucket = "tasks"

var db, _ = bolt.Open("db", 0600, nil)

func put(key string, value string) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			b, _ = tx.CreateBucket([]byte(bucket))
		}
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

func getall() [][]string {
	r := make([][]string, 0)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))

		b.ForEach(func(k, v []byte) error {
			r = append(r, []string{string(k), string(v)})
			return nil
		})

		return nil
	})

	return r
}

func main() {
	defer db.Close()

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
			put(task, "0")
			return c.String(http.StatusOK, "Success!\n")
		}
		return c.String(http.StatusOK, "Failed!\n")
	})

	e.POST("/list", func(c echo.Context) error {
		res := "Tasks:\n"
		r := getall()
		for i := range r {
			var status string
			if r[i][1] == "0" {
				status = "Not finished"
			} else {
				status = "Finished"
			}
			res += fmt.Sprintf("%v ----- %v\n", r[i][0], status)
		}
		return c.String(http.StatusOK, res)
	})

	e.Logger.Fatal(e.Start(":3333"))
}
