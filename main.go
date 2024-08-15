package main

import (
	"fmt"
	"net/http"

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

	e := echo.New()

	e.POST("/new", func(c echo.Context) error {
		task := c.FormValue("task")

		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucket))
			if b == nil {
				b, _ = tx.CreateBucket([]byte(bucket))
			}
			b.Put([]byte(task), []byte("0"))
			return nil
		})

		return c.String(http.StatusOK, "Test")
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
