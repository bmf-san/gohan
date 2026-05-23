package processor

import "testing"

func TestCountWords_English(t *testing.T) {
	got := countWords("Hello world foo bar")
	if got != 4 {
		t.Errorf("got %d, want 4", got)
	}
}

func TestCountWords_Japanese(t *testing.T) {
	// 5 hiragana characters → 5 words.
	got := countWords("こんにちは")
	if got != 5 {
		t.Errorf("got %d, want 5", got)
	}
}

func TestCountWords_Mixed(t *testing.T) {
	// "Hello 世界" → 1 English word + 2 CJK chars = 3.
	got := countWords("Hello 世界")
	if got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}

func TestCountWords_Empty(t *testing.T) {
	if got := countWords(""); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestReadingTimeMinutes(t *testing.T) {
	cases := []struct {
		words int
		want  int
	}{
		{0, 0},
		{1, 1},
		{249, 1},
		{250, 1},
		{251, 2},
		{500, 2},
		{501, 3},
	}
	for _, c := range cases {
		if got := readingTimeMinutes(c.words); got != c.want {
			t.Errorf("readingTimeMinutes(%d) = %d, want %d", c.words, got, c.want)
		}
	}
}
