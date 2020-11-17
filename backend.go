package main

import (
	"errors"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

const my_template = `{{define "index"}}<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>File download</title>
</head>
<body>
    <img src="1200.jpg" alt="Italian Trulli" width="500">

    
    <form action="/" method="post">
  <label for="todo">To Do :</label>
  <input type="text" id="todo" name="todo"><br>
	<input type="submit" value="Submit">
</form>
    
    <ul>
  {{range .Items}}<li>{{ . }}</li>{{else}}<li>no todo</li>{{end}}
</ul>

</body>
</html>{{end}}
`

var lastdownload time.Time
var conn *pgx.Conn

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func createtable() {
	_, err := conn.Exec(context.Background(), "CREATE TABLE todos (todo VARCHAR(100) NOT NULL);")
	check(err)
}

func check_table_exist() bool {
	var exist bool
	err := conn.QueryRow(context.Background(), "SELECT EXISTS (SELECT FROM information_schema.tables WHERE  table_name = 'todos');").Scan(&exist)
	check(err)
	return exist
}

func db_get_todo() []string {
	if !check_table_exist() {
		createtable()
		return []string{}
	}
	rows, _ := conn.Query(context.Background(), "select * from todos")

	todos := []string{}

	for rows.Next() {
		var todo string
		err := rows.Scan(&todo)
		check(err)
		todos = append(todos, todo)
	}

	check(rows.Err())
	return todos

}

func createTodo(c echo.Context) error {
	if !check_table_exist() {
		createtable()
	}

	todo := c.FormValue("todo")
	_, err := conn.Exec(context.Background(), "INSERT INTO todos (todo) VALUES($1);", todo)
	check(err)

	data := struct {
		Items []string
	}{
		Items: []string{},
	}
	data.Items = db_get_todo()

	return c.Render(http.StatusOK, "index", data)
}

func main() {

	var err error
	db_url := os.Getenv("DB_URL")
	if db_url == "" {
		db_url = "postgresql://postgres:test1234@joseph-test.chld9kh33qyg.us-east-1.rds.amazonaws.com:5432/postgres"
	}
	conn, err = pgx.Connect(context.Background(), db_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
		os.Exit(1)
	}

	err = downloadFile("https://picsum.photos/1200", "1200.jpg")
	check(err)

	lastdownload = time.Now()

	t := &Template{
		templates: template.Must(template.New("my-template").Parse(my_template)),
	}

	e := echo.New()
	e.Renderer = t

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		data := struct {
			Items []string
		}{
			Items: []string{},
		}
		data.Items = db_get_todo()
		return c.Render(http.StatusOK, "index", data)
	})
	e.GET("/1200.jpg", func(c echo.Context) error {
		now := time.Now()
		if now.Sub(lastdownload).Hours() >= 24 {
			err := downloadFile("https://picsum.photos/1200", "1200.jpg")
			check(err)
			lastdownload = now
		}
		return c.File("1200.jpg")
	})

	e.POST("/", createTodo)

	e.Logger.Fatal(e.Start(":9936"))
}
