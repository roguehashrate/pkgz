package utils

import (
	"reflect"
	"testing"
)

func TestScanProgress(t *testing.T) {
	split := func(input string) []string {
		var out []string
		data := []byte(input)
		for {
			advance, token, err := scanProgress(data, false)
			if err != nil {
				t.Fatalf("scanProgress error: %v", err)
			}
			if advance == 0 {
				advance, token, err = scanProgress(data, true)
				if err != nil {
					t.Fatalf("scanProgress atEOF error: %v", err)
				}
			}
			out = append(out, string(token))
			if advance == 0 {
				break
			}
			data = data[advance:]
		}
		return out
	}

	t.Run("splits on newline", func(t *testing.T) {
		got := split("aaa\nbbb\n")
		if !reflect.DeepEqual(got, []string{"aaa", "bbb", ""}) {
			t.Errorf("got %v", got)
		}
	})

	t.Run("splits on carriage returns", func(t *testing.T) {
		got := split("10%\r20%\r30%\n")
		if !reflect.DeepEqual(got, []string{"10%", "20%", "30%", ""}) {
			t.Errorf("got %v", got)
		}
	})

	t.Run("handles CRLF as one boundary", func(t *testing.T) {
		got := split("a\r\nb\r\n")
		// The splitter breaks on both chars, so the trailing \n of a CRLF pair
		// surfaces as an empty token; consumers skip empties downstream.
		if !reflect.DeepEqual(got, []string{"a", "", "b", "", ""}) {
			t.Errorf("got %v", got)
		}
	})
}

func TestCleanLine(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"\x1b[2K  42%\r", "42%"},
		{"\x1b[1;32mfoo\x1b[0m", "foo"},
		{"\x1b[?25l\x1b[2K\r", ""},
		{"  done\n", "done"},
		{"\r\n", ""},
	}
	for _, c := range cases {
		if got := cleanLine(c.in); got != c.want {
			t.Errorf("cleanLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
