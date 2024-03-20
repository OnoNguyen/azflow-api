// use the main package
package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	// create a new instance of the Fiber app
	app := fiber.New()

	// define a route
	app.Get("/:value", func(c *fiber.Ctx) error {
		return c.SendString("the best value is: " + c.Params("value"))
	})

	// start the server on port 3000
	log.Fatal(app.Listen("localhost:4000"))
}
