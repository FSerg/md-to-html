package ui

import (
	"bytes"
	"context"
	"testing"
)

func TestHomeRenderSmoke(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := Home().Render(context.Background(), &buf); err != nil {
		t.Fatalf("render home: %v", err)
	}

	if got := buf.Len(); got <= 500 {
		t.Fatalf("rendered output too small: %d", got)
	}
}
