package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/GLVSKiriti/URLshortner/database"
	"github.com/GLVSKiriti/URLshortner/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
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

	r2 := database.CreateClient(1)
	defer r2.Close()

	val, err := r2.Get(database.Ctx, c.IP()).Result()
	//This happens when the given ip address user is not exits in databse it sets the new key value pair now
	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		//Happens when ip address already exists in database so now it checks rate limit now
		// val, _ := r2.Get(database.Ctx, c.IP()).Result()
		//You can also no need to write above line as already we rote the same line above so we can use that val
		valint, _ := strconv.Atoi(val)
		if valint <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":            "rate limit exceeded",
				"rate-limit-reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	//check if the input sent by user is actual URL or not
	if !govalidator.IsURL(body.URL) {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	//check for domain error
	//It checks whether the user is trying to hack or manipulate the our website
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "You can't hack the system"})
	}

	//enforce https,SSL
	body.URL = helpers.EnforceHTTP(body.URL)

	//Here the below code does like
	//it checks whether custom url given by user is exists in database or not
	//if not it sets the key value in database
	//if no custom url is given by the user then random url is generated
	//Ans set the expiry time here
	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "URL custom short is already in use",
		})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to connect to server",
		})
	}

	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	//This line at last decrements the ratelimit by 1
	r2.Decr(database.Ctx, c.IP())

	val, _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
