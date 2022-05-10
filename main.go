package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"io"
	"log"
	"os"
	"strings"
)

var ResourcePath = map[string]string{}
var ResourceData = map[string][]byte{}
var ResourceHash = map[string]string{}

func init() {
	open, err := os.Open("resources.json")
	if err != nil {
		log.Fatalf("failed to open resource: %s", err)
	}

	data, err := io.ReadAll(open)
	if err != nil {
		log.Fatalf("failed to read resource: %s", err)
	}

	err = json.Unmarshal(data, &ResourcePath)
	if err != nil {
		log.Fatalf("failed to unmarshal json data: %s", err)
	}

	for k, v := range ResourcePath {
		open, err = os.Open("Resources/" + v)
		if err != nil {
			log.Fatalf("failed to open resource: %s", err)
		}

		data, err = io.ReadAll(open)
		if err != nil {
			log.Fatalf("failed to read resource: %s", err)
		}

		ResourceData[k] = data
		ResourceHash[k] = fmt.Sprintf("%x", sha256.Sum256(data))
	}
}

func main() {
	app := fiber.New(fiber.Config{
		Prefork: true,
	})

	app.Use(recover.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Get("/:type/:hash", downloadResource)

	log.Fatal(app.Listen(":3001"))
}

func downloadResource(c *fiber.Ctx) error {
	c.Set("surrogate-key", "mod-resource")

	resourceType := strings.ToLower(c.Params("type"))
	resourceData := ResourceData[resourceType]

	if resourceData == nil {
		return c.SendStatus(404)
	}

	resourceHash := strings.ToLower(c.Params("hash"))

	if resourceHash == ResourceHash[resourceType] {
		return c.SendStatus(204)
	}

	c.Set("content-type", "application/octet-stream")

	return c.Send(resourceData)
}
