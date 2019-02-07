package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommonLogFormatParser(t *testing.T) {
	p := NewParser()

	t.Run("Parse well-formed line succeeds", func(t *testing.T) {
		line := "127.0.0.1 - james [09/May/2018:16:00:39 +0000] \"GET /report HTTP/1.0\" 200 123"

		r, err := p.ParseLine(line)
		assert.NoError(t, err)

		// remote assertions
		assert.Equal(t, "127.0.0.1", r.RemoteHost)
		assert.Equal(t, "-", r.RemoteLogname)
		assert.Equal(t, "james", r.AuthUser)

		// datetime assertions
		assert.Equal(t, "May", r.Timestamp.Month().String())
		assert.Equal(t, 9, r.Timestamp.Day())
		assert.Equal(t, 2018, r.Timestamp.Year())
		assert.Equal(t, 16, r.Timestamp.Hour())
		assert.Equal(t, 00, r.Timestamp.Minute())
		assert.Equal(t, 39, r.Timestamp.Second())
		zoneName, zoneOffset := r.Timestamp.Zone()
		assert.Equal(t, "", zoneName)
		assert.Equal(t, 0, zoneOffset)

		// request assetions
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/report", r.URI)
		assert.Equal(t, "1.0", r.HTTPVersion)
		assert.Equal(t, 200, r.StatusCode)
		assert.Equal(t, 123, r.ResponseSizeBytes)
	})

	t.Run("Parse empty line returns error", func(t *testing.T) {
		r, err := p.ParseLine("")
		assert.Nil(t, r)
		assert.Error(t, err)
	})

	t.Run("Parse line with nil response size", func(t *testing.T) {
		line := "127.0.0.1 - james [09/May/2018:16:00:39 +0000] \"GET /report HTTP/1.0\" 200 -"
		r, err := p.ParseLine(line)
		assert.NoError(t, err)

		assert.Equal(t, 0, r.ResponseSizeBytes)
	})
}
