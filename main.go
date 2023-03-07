package main

import (
	"errors"
	"fmt"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/sony/gobreaker"
) 

var startTime time.Time = time.Now()

func main() {
	
	go server()
	
	cb := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name: "my-circuit-breaker",
			MaxRequests: 5,
			Timeout: 3 * time.Second,
			Interval: 1 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 3
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				fmt.Printf("circuitbreaker '%s' changed from '%s' to '%s'\n", name, from, to)
			},
		},
	)

	fmt.Println("call with curcuitbreaker")
	for i := 0; i < 100; i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			err := DoReq()
			return nil, err
		})
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
    
}

func server() {
	
	app := fiber.New()
	app.Use(logger.New())

    app.Get("/ping", func(c *fiber.Ctx) error {
        if time.Since(startTime) < 5*time.Second {
			return fiber.NewError(fiber.StatusInternalServerError, "pong")
		}
		c.SendStatus(fiber.StatusOK)
		return c.SendString("pong")
    })

	fmt.Println("Starting server at port 8080")
    app.Listen(":3000")
}

func DoReq() error {
	a := fiber.AcquireAgent()
	req := a.Request()
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI("http://127.0.0.1:3000/ping")
	
	if err := a.Parse(); err != nil {
		panic(err)
	}
	
	code, body, errs := a.String()
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err)
		}
		return errors.New("bad response")
	}
	
	if code < 200 || code >= 300 {
		return errors.New("bad response")
	}
	fmt.Println(body)
	return nil

}