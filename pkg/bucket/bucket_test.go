package bucket

import (
	"testing"
)

// TestBucket
func TestBucket(t *testing.T) {
	_, err := New(SetName("name"), SetRegion("eu-west-2"))
	if err != nil {
		t.Log(err)
	}
}
