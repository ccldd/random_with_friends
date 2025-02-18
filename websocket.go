package main

type WSMessageType string

const (
	WSMessageTypeError WSMessageType = "error"
)

type WSMessage struct {
	Type string `json:"type"`
}

type WSMessageError struct {
	WSMessage
	Error string `json:"error"`
}

func NewWSMessageError(err error) *WSMessageError {
	return &WSMessageError{
		WSMessage: WSMessage{Type: string(WSMessageTypeError)},
		Error:     err.Error(),
	}
}