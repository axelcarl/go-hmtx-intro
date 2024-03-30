package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"strconv"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

var id int = 0

type Contact struct {
	Name  string
	Email string
	Id    int
}

func newContact(name, email string) Contact {
	id++
	return Contact{
		Name:  name,
		Email: email,
		Id:    id,
	}
}

type Contacts = []Contact

type Data struct {
	Contacts Contacts
}

func (d Data) hasEmail(email string) bool {
	for _, contact := range d.Contacts {
		if email == contact.Email {
			return true
		}
	}

	return false
}

func (d *Data) indexOf(id int) int {
	for index, contact := range d.Contacts {
		if contact.Id == id {
			return index
		}
	}

	return -1
}

func newData() Data {
	return Data{
		Contacts: Contacts{
			newContact("Charlie", "c@harvard.com"),
			newContact("Ross", "r@mit.edu"),
		},
	}
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	Data Data
	Form FormData
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func main() {

	e := echo.New()
	e.Use(middleware.Logger())

	page := newPage()
	e.Renderer = newTemplate()

	// Serve static content
	e.Static("/images", "images")
	e.Static("/css", "css")

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.hasEmail(email) {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists"

			return c.Render(422, "form", formData)
		}
		contact := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contact)

		c.Render(200, "form", newFormData())
		return c.Render(200, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {

		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)

		if err != nil {

			return c.String(400, "Invalid Id")
		}

		index := page.Data.indexOf(id)

		if index == -1 {
			return c.String(400, "Contact not found")
		}

		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)

		return c.NoContent(200)
	})

	e.Logger.Fatal(e.Start(":42069"))
}
