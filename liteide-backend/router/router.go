package router

import (
	"errors"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"message": err.Error(),
	})
}

func useWS(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func usePagination() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page, err := strconv.ParseInt(c.Query("page", "1"), 10, 64)
		if page <= 0 || err != nil {
			page = 1
		}
		size, err := strconv.ParseInt(c.Query("size", "10"), 10, 64)
		if size < 0 || err != nil {
			size = 10
		}

		c.Locals("offset", int((page-1)*size))
		c.Locals("limit", int(size))

		return c.Next()
	}
}
