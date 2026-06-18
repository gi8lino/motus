package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("boom")
}

func TestNewRequestID(t *testing.T) {
	t.Run("generates a hex id", func(t *testing.T) {
		t.Parallel()

		id := NewRequestID()
		assert.Len(t, id, 32)
	})

	t.Run("returns empty string on rand failure", func(t *testing.T) {
		// Cannot run in parallel because it mutates crypto/rand.Reader.
		old := requestIDReader
		requestIDReader = errReader{}
		defer func() { requestIDReader = old }()

		id := NewRequestID()
		assert.Equal(t, "", id)
	})
}

func TestRequestID(t *testing.T) {
	t.Parallel()

	t.Run("returns empty on nil context", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "", RequestID(context.TODO()))
	})

	t.Run("returns stored id", func(t *testing.T) {
		t.Parallel()

		ctx := WithRequestID(context.Background(), "req-123")
		assert.Equal(t, "req-123", RequestID(ctx))
	})
}

func TestWithRequestIDLogger(t *testing.T) {
	t.Parallel()

	t.Run("adds request_id when present", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		logger := slog.New(slog.NewJSONHandler(buf, nil))
		ctx := WithRequestID(context.Background(), "req-456")

		WithRequestIDLogger(logger, ctx).Info("hello")

		entry := readLogJSON(t, buf)
		assert.Equal(t, "req-456", entry["request_id"])
	})

	t.Run("does not panic with nil logger", func(t *testing.T) {
		t.Parallel()

		require.NotPanics(t, func() {
			WithRequestIDLogger(nil, context.Background()).Info("hello")
		})
	})
}

func readLogJSON(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	line := strings.TrimSpace(buf.String())
	if idx := strings.IndexByte(line, '\n'); idx != -1 {
		line = line[:idx]
	}

	var entry map[string]any
	err := json.Unmarshal([]byte(line), &entry)
	require.NoError(t, err)
	return entry
}
