package handlers

import "testing"

func TestScp(t *testing.T) {
	// t.Log(scp([]string{"../testdata/example.ipa", ":/tmp/example.ipa"}))
	t.Log(scp([]string{":/tmp/example.ipa", "../testdata/test.ipa"}))
}
