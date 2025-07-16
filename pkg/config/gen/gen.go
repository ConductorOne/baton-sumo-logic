package main

import (
	"github.com/conductorone/baton-sdk/pkg/config"
	cfg "github.com/conductorone/baton-sumo-logic/pkg/config"
)

func main() {
	config.Generate("sumo-logic", cfg.Config)
}
