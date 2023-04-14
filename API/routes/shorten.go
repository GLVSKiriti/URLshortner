package routes

import (
	"time"

	"github.com/GLVSKiriti/URLshortner/helpers"
	"github.com/gofiber/fiber/v2"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaining int           `json:"rate_limit"`
	XRateLimitRest time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {

	//here we are making a struct type of request and named it as body
	body := new(request)

	//this checks whether it is valid request or not like it parses the request
	//as we are writing like &body (referencing) it parses and stores in body only
	//if there is no error
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot Parse the JSON"})
	}

	//implementing rate limiting

	//check if the input sent by user is actual URL or not
	if !govalidator.IsUrl(body.URL) {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	//check for domain error
	//It checks whether the user is trying to hack or manipulate the our website
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "You can't hack the system"})
	}

	//enforce https,SSL
	body.URL = helpers.EnforceHTTP(body.URL)

}
