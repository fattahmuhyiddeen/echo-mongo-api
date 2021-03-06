package main

import (
	config "./config"
	"./handler"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"gopkg.in/mgo.v2"
)

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.ERROR)
	e.Use(middleware.Logger())
	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(handler.Key),
		Skipper: func(c echo.Context) bool {
			// Skip authentication for and signup login requests
			if c.Path() == "/login" || c.Path() == "/signup" || c.Path() == "/verify" || c.Path() == "/forget" || c.Path() == "/test" {
				return true
			}
			return false
		},
	}))

	// Database connection
	db, err := mgo.Dial("localhost")
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Create indices
	if err = db.Copy().DB(config.DbName).C("users").EnsureIndex(mgo.Index{
		Key:    []string{"email"},
		Unique: true,
	}); err != nil {
		log.Fatal(err)
	}

	// Initialize handler
	h := &handler.Handler{DB: db}

	// Routes
	e.POST("/signup", h.Signup)
	e.POST("/login", h.Login)
	e.GET("/verify", h.Verify)
	// e.POST("/follow/:id", h.Follow)
	// e.POST("/posts", h.CreatePost)
	// e.GET("/feed", h.FetchPost)
	e.GET("/profile", h.GetProfile)
	e.POST("/profile", h.UpdateProfile)
	e.POST("/password", h.UpdatePassword)
	e.GET("/forget", h.RequestChangePassword)
	e.POST("/forget", h.ResetPassword)
	e.GET("/test", h.TestFunc)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
