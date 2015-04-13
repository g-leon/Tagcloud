package main

import "testing"

func TestIsWord(t *testing.T) {
	cases := []string{
		"a",
		"http://website.com",
		"https://website.com",
		"@user",
		"1234",
	}

	for _, c := range cases {
		if isWord(c) {
			t.Errorf("Expected false on %q", c)
		}
	}
}

func TestTrimWord(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{" word ", "word"},
		{"!,.?;!$%^&*()[]{}'/|><~`+-=\\\"word!,.?;!$%^&*()[]{}'/|><~`+-=\\\"", "word"},
		{" \t\n word \n\t\r\n ", "word"},
	}

	for _, c := range cases {
		got := trimWord(c.in)
		if got != c.want {
			t.Errorf("Expected %q and got %q on word %q", c.want, got, c.in)
		}
	}
}

func TestSplitText(t *testing.T) {
	in1 := "a b"
	in2 := "a\nb"
	want := []string{"a", "b"}

	i := 0
	for _, v := range splitText(in1) {
		if v != want[i] {
			t.Errorf("Expected %q and got %q", want[i], v)
		}
		i++
	}

	i = 0
	for _, v := range splitText(in2) {
		if v != want[i] {
			t.Errorf("Expected %q and got %q", want[i], v)
		}
		i++
	}
}
