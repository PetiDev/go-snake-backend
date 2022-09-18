package main

import (
	"context"
	"os"
	"time"

	"snakeLeaderboard/db"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html"
	"golang.org/x/crypto/bcrypt"
)

type Request struct {
	Name         string `json:"name"`
	Password     string `json:"password"`
	Points       int    `json:"points"`
	Achievements int    `json:"achievements"`
}
type Data struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Pass         string `json:"pass"`
	Points       int    `json:"points"`
	Achievements int    `json:"achievements"`
}

func main() {
	gameVerison := os.Getenv("GAMEVERSION")
	db.Connect()
	defer db.Disconnect()

	app := fiber.New(fiber.Config{
		Views:        html.New("./src", ".html"),
		Prefork:      true,
		ServerHeader: "Never gonna give you up, Never gonna let you down",
		AppName:      "Backend for snake leaderboard",
	})
	app.Use(favicon.New(favicon.Config{
		File: "./src/snake.ico",
	}))

	ctx := context.Background()

	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./src/index.html")
	})
	app.Get("/get", func(c *fiber.Ctx) error {
		res, err := db.DB.Test.FindMany().OrderBy(db.Test.Points.Order(db.DESC)).Exec(ctx)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		for x := range res {
			res[x].Pass = "never gonna give you up, never gonna let you down."
		}
		return c.JSON(res)

	})

	app.Get("/get/version", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusTeapot).SendString(gameVerison)
	})
	app.Post("/register", func(c *fiber.Ctx) error {
		if c.Query("version") != gameVerison {
			return c.Status(fiber.StatusTeapot).SendString("Outdated game version")
		}
		data := new(Request)
		if c.BodyParser(data) != nil {
			return c.Status(400).SendString("Sikertelen cucc")
		}

		res, _ := db.DB.Test.FindFirst(
			db.Test.Name.Equals(data.Name),
		).Exec(ctx)

		if res != nil {
			return c.Status(fiber.StatusConflict).SendString("Már létezik a felhasználó")
		}
		hash, err := bcrypt.GenerateFromPassword(
			[]byte(data.Password),
			12,
		)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		_, err = db.DB.Test.CreateOne(
			db.Test.Name.Set(data.Name),
			db.Test.Pass.Set(string(hash)),
			db.Test.Points.Set(data.Points),
			db.Test.Achievements.Set(data.Achievements),
		).Exec(ctx)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.Status(fiber.StatusCreated).SendString("Sikires regisztráció")
	})
	app.Post("/write", func(c *fiber.Ctx) error {
		if c.Query("version") != gameVerison {
			return c.Status(fiber.StatusTeapot).SendString("Outdated game version")
		}
		data := new(Request)
		if c.BodyParser(data) != nil {
			return c.Status(400).SendString("Sikertelen cucc")
		}

		res, err := db.DB.Test.FindFirst(
			db.Test.Name.Equals(data.Name),
		).Exec(ctx)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		err = bcrypt.CompareHashAndPassword(
			[]byte(res.Pass),
			[]byte(data.Password),
		)
		if err != nil {
			return c.Status(fiber.StatusForbidden).SendString(err.Error())
		}
		if data.Points <= res.Points {
			return c.Status(fiber.StatusNotModified).SendString("A pillanatnyilag elért pontszám nem nagyobb mint a nyilvántartásban szereplő legmagasabb eddigi elért pont ezen a felhasználón")
		}
		_, err = db.DB.Test.FindMany(
			db.Test.ID.Equals(res.ID),
		).Update(
			db.Test.Points.Set(data.Points),
		).Exec(ctx)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.Status(fiber.StatusOK).SendString("A rekord sikeresen felülírásra került az adatbázisban a pillanatynyilag elért legmagasabb pontszámra")

	})

	app.Get("/delete/:id", func(c *fiber.Ctx) error {
		if c.Query("key") != os.Getenv("API_KEY") {
			return c.Status(fiber.StatusForbidden).SendString("Incorrect key")
		}
		_, err := db.DB.Test.FindMany(
			db.Test.ID.Equals(c.Params("id")),
		).Delete().Exec(ctx)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		return c.Status(fiber.StatusOK).SendString("Successfully deleted")
	})

	// Or extend your config for customization
	app.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("refresh") == "true"
		},
		Expiration:   5 * time.Minute,
		CacheControl: true,
	}))

	app.Get("/scoreboard", func(c *fiber.Ctx) error {
		res, err := db.DB.Test.FindMany().OrderBy(db.Test.Points.Order(db.DESC)).Exec(ctx)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.Render("scoreboard", res)

	})

	app.Listen(os.Getenv("HOST"))
}
