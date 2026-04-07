package service

import "testing"

func TestBaiduShorturl_FromShareURL(t *testing.T) {
	in := "https://pan.baidu.com/s/1abcdefghijklmnopqrstuvwxyz"
	got := baiduShorturl(in)
	if got != "1abcdefghijklmnopqrstuvwxyz" {
		t.Fatalf("baiduShorturl() got=%q want=%q", got, "1abcdefghijklmnopqrstuvwxyz")
	}
}

func TestBaiduShorturlCandidates_WithLeadingOneFallback(t *testing.T) {
	in := "https://pan.baidu.com/s/1abcdefghijklmnopqrstuvwxyz"
	got := baiduShorturlCandidates(in)
	if len(got) != 2 {
		t.Fatalf("len(candidates)=%d want=2", len(got))
	}
	if got[0] != "1abcdefghijklmnopqrstuvwxyz" || got[1] != "abcdefghijklmnopqrstuvwxyz" {
		t.Fatalf("candidates=%v", got)
	}
}
