package common

import "errors"

var ErrVideoNotFound = errors.New("video not found")

var ErrVideoAlreadyAdded = errors.New("This video already exists in your database")

var ErrUserDoesntExist = errors.New("The specified user wasn't found")
