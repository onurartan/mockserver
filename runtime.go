package main

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	msconfig "mockserver/config"
)

type Runtime struct {
	App    *fiber.App
	Cfg    *msconfig.Config
	Mu     sync.Mutex
}
