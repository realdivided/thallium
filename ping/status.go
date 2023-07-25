package ping

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/Tnze/go-mc/chat"
	_ "github.com/Tnze/go-mc/data/lang/en-us"
	"github.com/google/uuid"
)

type Status struct {
	Description chat.Message
	Latency     time.Duration
	Players     StatusPlayers
	Version     StatusVersion
	Favicon     Icon
}

type StatusPlayers struct {
	Max    int
	Online int
	Sample []struct {
		ID   uuid.UUID
		Name string
	}
}

type StatusVersion struct {
	Name     string
	Protocol int
}

type Icon string

var IconFormatErr = errors.New("data format error")
var IconAbsentErr = errors.New("icon not present")

// ToPNG decode base64-icon, return a PNG image
// Take care of there is not safety check, image may contain malicious code .
func (i Icon) ToPNG() ([]byte, error) {
	const prefix = "data:image/png;base64,"
	if i == "" {
		return nil, IconAbsentErr
	}
	if !strings.HasPrefix(string(i), prefix) {
		return nil, IconFormatErr
	}
	return base64.StdEncoding.DecodeString(strings.TrimPrefix(string(i), prefix))
}
