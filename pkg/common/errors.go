package common

import "errors"

var ErrVideoNotFound = errors.New("video not found")

var ErrVideoAlreadyAdded = errors.New("this video already exists in your database")

var ErrUserDoesNotExist = errors.New("the specified user wasn't found")
