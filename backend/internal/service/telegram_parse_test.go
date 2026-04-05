package service

import (
	"strings"
	"testing"
)

func TestParseTGResourceContentImgSrcCover(t *testing.T) {
	html := `<p></p><div class="tgme_widget_message_text">游戏介绍<br><a href="https://pan.baidu.com/s/1x">百度</a> <a href="https://pan.quark.cn/s/y">夸克</a></div><p></p><img src="https://cdn4.telesco.pe/file/WF41FIRST.jpg" width="453"><img src="https://cdn4.telesco.pe/file/SECOND.jpg" width="149">`
	got := parseTGResourceContent(html)
	if !strings.HasPrefix(got.Cover, "https://cdn4.telesco.pe/file/WF41FIRST") {
		t.Fatalf("cover = %q", got.Cover)
	}
	if got.Link == "" || !strings.Contains(got.Link, "pan.baidu.com") {
		t.Fatalf("link = %q", got.Link)
	}
}
