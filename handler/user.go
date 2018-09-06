package handler

import (
	"net/http"
	"time"

	config "../config"
	logic "../logic"

	"../model"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (h *Handler) Signup(c echo.Context) (err error) {
	// Bind
	u := &model.User{ID: bson.NewObjectId()}
	if err = c.Bind(u); err != nil {
		return
	}

	// Validate
	if !logic.IsValidEmail(u.Email) {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid email"}
	}
	if !logic.IsValidPassword(u.Password) {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid password"}
	}

	db := h.DB.Clone()
	defer db.Close()

	numRows, err := db.DB(config.DbName).C("users").Find(bson.M{"email": u.Email}).Count()

	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Sorry, please try later"}
	}
	if numRows > 0 {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Email already taken, please choose another email"}
	}

	if err = db.DB(config.DbName).C("users").Insert(u); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Sorry, please try later"}
	}

	verifyKey, err := logic.GenerateRandomStringURLSafe(32)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Sorry, verify key error"}
	}

	if err = db.DB(config.DbName).C("users").
		UpdateId(u.ID, bson.M{"$set": bson.M{"verifyKey": verifyKey, "status": "Unverified", "password": logic.HashPassword(u.Password)}}); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
	}
	u.Password = ""

	return c.JSON(http.StatusCreated, u)
}

func (h *Handler) Login(c echo.Context) (err error) {
	// Bind
	u := new(model.User)
	if err = c.Bind(u); err != nil {
		return
	}

	// Find user
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB(config.DbName).C("users").
		// Find(bson.M{"email": u.Email, "password": u.Password}).One(u); err != nil {
		Find(bson.M{"email": u.Email}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid email"}
		}
		return
	}

	if !logic.ComparePasswords(u.Password, c.FormValue("password")) {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid password"}
	}

	//-----
	// JWT
	//-----

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = u.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response
	u.Token, err = token.SignedString([]byte(Key))
	if err != nil {
		return err
	}

	u.Password = "" // Don't send password
	return c.JSON(http.StatusOK, u)
}

//Verify to verify user (clicked by user in email)
func (h *Handler) Verify(c echo.Context) (err error) {
	// Bind
	u := new(model.User)
	if err = c.Bind(u); err != nil {
		return
	}

	email := c.QueryParam("e")
	verifyToken := c.QueryParam("t")

	// Find user
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB(config.DbName).C("users").
		Find(bson.M{"email": email, "verifyToken": verifyToken}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid link or token already expired"}
		}
		return
	}

	if err = db.DB(config.DbName).C("users").
		UpdateId(u.ID, bson.M{"$set": bson.M{"verifyToken": ""}}); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
	}

	return c.JSON(http.StatusOK, "ok")
}

// func (h *Handler) Follow(c echo.Context) (err error) {
// 	userID := userIDFromToken(c)
// 	id := c.Param("id")

// 	// Add a follower to user
// 	db := h.DB.Clone()
// 	defer db.Close()
// 	if err = db.DB(config.DbName).C("users").
// 		UpdateId(bson.ObjectIdHex(id), bson.M{"$addToSet": bson.M{"followers": userID}}); err != nil {
// 		if err == mgo.ErrNotFound {
// 			return echo.ErrNotFound
// 		}
// 	}

// 	return
// }

// GetProfile to profile of the user
func (h *Handler) GetProfile(c echo.Context) (err error) {
	userID := userIDFromToken(c)

	// Retrieve posts from database
	user := model.User{}
	db := h.DB.Clone()

	if err = db.DB(config.DbName).C("users").FindId(bson.ObjectIdHex(userID)).One(&user); err != nil {
		return
	}

	defer db.Close()

	user.Password = ""

	return c.JSON(http.StatusOK, user)
}

// UpdateProfile to update profile of the user
func (h *Handler) UpdateProfile(c echo.Context) (err error) {
	userID := userIDFromToken(c)
	name := c.FormValue("name")

	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB(config.DbName).C("users").
		UpdateId(bson.ObjectIdHex(userID), bson.M{"$set": bson.M{"name": name}}); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
	}

	return c.JSON(http.StatusOK, "ok")
}

// UpdatePassword to update profile of the user
func (h *Handler) UpdatePassword(c echo.Context) (err error) {
	userID := userIDFromToken(c)
	newPassword := c.FormValue("newPassword")
	oldPassword := c.FormValue("oldPassword")

	user := model.User{}
	db := h.DB.Clone()

	if err = db.DB(config.DbName).C("users").FindId(bson.ObjectIdHex(userID)).One(&user); err != nil {
		return
	}
	defer db.Close()

	if !logic.ComparePasswords(user.Password, oldPassword) {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "wrong old password"}
	}

	if !logic.IsValidPassword(newPassword) {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid new password"}
	}

	if err = db.DB(config.DbName).C("users").
		UpdateId(bson.ObjectIdHex(userID), bson.M{"$set": bson.M{"password": logic.HashPassword(newPassword)}}); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
	}

	return
}

func userIDFromToken(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["id"].(string)
}
