package main

import (
	"errors"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"io"
	"net/http"
	"os"
	"time"
)

var lastdownload time.Time

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

func main() {

	err := downloadFile("https://picsum.photos/1200", "1200.jpg")
	check(err)

	lastdownload = time.Now()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.File("index.html")
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

	e.Logger.Fatal(e.Start(":9936"))
}
