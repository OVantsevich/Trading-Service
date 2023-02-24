package repository

import (
	"os"
	"testing"
)

var listenersRepository *ListenersRepository

func TestMain(m *testing.M) {
	listenersRepository = NewListenersRepository()
	code := m.Run()
	os.Exit(code)
}
