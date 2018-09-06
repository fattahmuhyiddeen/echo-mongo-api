package handler

import (
	"net/http"

	"github.com/labstack/echo"
)

// TestFunc to profile of the user
func (h *Handler) TestFunc(c echo.Context) (err error) {
	name := c.QueryParam("name")

	return c.JSON(http.StatusOK, name)
}
