package main

import (
	"testing"
	"time"
)

func TestFileNameProcessor_Generate(t *testing.T) {
	out := NewProcessor().Generate("test", time.Now())
	t.Logf("result %v", out)
}

func TestNeedDeleteFile(t *testing.T) {
	out := NewProcessor().Generate("test_cc_s", time.Now().AddDate(0, 0, -9))
	t.Logf("Generate %v", out)
	res := NeedDeleteFile("test_cc_s", out)
	t.Logf("NeedDeleteFile %v", res)
}
