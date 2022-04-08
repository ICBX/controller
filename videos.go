package main

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
)

func addNewVideo(ctx *fiber.Ctx) error {

	req := NewVideoPayload{}

	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return errors.New("Could not parse body: " + err.Error())
	}

	// TODO: add new video to download queue

	return nil
}
