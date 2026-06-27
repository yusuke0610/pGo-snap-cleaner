package findertag

import (
	"os"
	"path/filepath"
	"testing"
)

func tempFile(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "img.jpg")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestAddPGoIdempotent(t *testing.T) {
	p := tempFile(t)

	added, err := AddPGo(p)
	if err != nil {
		t.Fatal(err)
	}
	if !added {
		t.Fatal("first AddPGo should report added=true")
	}

	// 2 回目は何もしないはず。
	added, err = AddPGo(p)
	if err != nil {
		t.Fatal(err)
	}
	if added {
		t.Fatal("second AddPGo should report added=false (idempotent)")
	}

	tags, err := ReadTags(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := countPGo(tags); got != 1 {
		t.Fatalf("expected exactly 1 pGo tag, got %d (%v)", got, tags)
	}
	if tags[indexOfPGo(tags)] != pGoTag {
		t.Fatalf("pGo tag not encoded as %q: %v", pGoTag, tags)
	}
}

func TestRemovePGoPreservesOtherTags(t *testing.T) {
	p := tempFile(t)

	// ユーザーが手動で付けた赤タグと緑タグが既にある状態。
	existing := []string{"Important\n6", "Work\n2"}
	if err := writeTags(p, existing); err != nil {
		t.Fatal(err)
	}
	if _, err := AddPGo(p); err != nil {
		t.Fatal(err)
	}

	removed, err := RemovePGo(p)
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatal("RemovePGo should report removed=true")
	}

	tags, err := ReadTags(p)
	if err != nil {
		t.Fatal(err)
	}
	if countPGo(tags) != 0 {
		t.Fatalf("pGo tag should be gone, got %v", tags)
	}
	if len(tags) != 2 {
		t.Fatalf("user tags must survive untag, got %v", tags)
	}
	want := map[string]bool{"Important\n6": true, "Work\n2": true}
	for _, tg := range tags {
		if !want[tg] {
			t.Fatalf("unexpected tag after untag: %q", tg)
		}
	}
}

func TestRemovePGoOnlyTagDropsAttr(t *testing.T) {
	p := tempFile(t)
	if _, err := AddPGo(p); err != nil {
		t.Fatal(err)
	}
	if _, err := RemovePGo(p); err != nil {
		t.Fatal(err)
	}
	// 属性ごと消えているはず。ReadTags は空を返し、エラーは出ない。
	tags, err := ReadTags(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 0 {
		t.Fatalf("expected no tags, got %v", tags)
	}
}

func TestRemovePGoNoopWhenAbsent(t *testing.T) {
	p := tempFile(t)
	removed, err := RemovePGo(p)
	if err != nil {
		t.Fatal(err)
	}
	if removed {
		t.Fatal("RemovePGo on a file with no tags should report removed=false")
	}
}

func countPGo(tags []string) int {
	n := 0
	for _, t := range tags {
		if tagName(t) == TagName {
			n++
		}
	}
	return n
}
