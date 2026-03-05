package requesthandler

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCheckURLs(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	assertCorrectMessage := func(t *testing.T, got, want int) {
		if got != want {
			t.Errorf("got %d calls, want %d calls", got, want)
		}
	}

	t.Run("invalid urls", func(t *testing.T) {

		urls := []string{
			"a",
			"b",
			"c",
			"",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		getter := &SpyGetter{}

		got := CheckURLs(ctx, urls, getter)
		want := map[string]Result{
			"a": {StatusCode: 0, Err: errors.New("Invalid url")},
			"b": {StatusCode: 0, Err: errors.New("Invalid url")},
			"c": {StatusCode: 0, Err: errors.New("Invalid url")},
			"":  {StatusCode: 0, Err: errors.New("Invalid url")},
		}

		if !equalMaps(t, got, want) {
			t.Errorf("got %v, want %v", got, want)
		}

		// // since we are giving invalid urls, we expect 0 calls

	})

	t.Run("valid URLs", func(t *testing.T) {

		urls := []string{
			"https://gmail.com/",
			"https://quii.gitbook.io/learn-go-with-tests",
			"https://google.com/",
			"https://leetcode.com/problemset/",
		}

		getter := &SpyGetter{}

		CheckURLs(ctx, urls, getter)

		want := 4

		assertCorrectMessage(t, getter.Calls, want)

	})

	t.Run("Final Correct Check", func(t *testing.T) {

		urls := []string{
			"https://gmail.com/",
			"https://quii.gitbook.io/learn-go-with-tests",
			"https://google.com/",
			"https://leetcode.com/problemset/",
		}

		getter := &SpyGetter{}

		got := CheckURLs(ctx, urls, getter)
		want := map[string]Result{
			"https://gmail.com/":                          {StatusCode: 200},
			"https://quii.gitbook.io/learn-go-with-tests": {StatusCode: 200},
			"https://google.com/":                         {StatusCode: 200},
			"https://leetcode.com/problemset/":            {StatusCode: 200},
		}

		assertCorrectMessage(t, getter.Calls, 4)

		if !equalMaps(t, got, want) {
			t.Errorf("got %v, want %v", got, want)
		}

	})

}

func equalMaps(t *testing.T, m1, m2 map[string]Result) bool {

	t.Helper()

	if len(m1) != len(m2) {
		return false
	}

	for k, v1 := range m1 {
		v2, _ := m2[k]
		if !errors.Is(v1.Err, v2.Err) {
			return false
		}
	}

	return true
}
