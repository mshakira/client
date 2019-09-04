package incidents_test

import (
	"client/incidents"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var incidentsObj *incidents.Incidents

func TestMain(m *testing.M) {
	var err error
	incidentsObj, err = incidents.Init()
	if err != nil {
		log.Printf("%s Failed, err:%s\n", "incidents.Init", err)
		os.Exit(1)
	}
	var status int = m.Run()

	os.Exit(status)
}

func TestParseBody(t *testing.T) {
	// failure case
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "<html><body>Hello World!</body></html>")
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	err := incidentsObj.ParseBody(resp)
	if err == nil {
		t.Errorf("Expected error, got no error")
	}
	if incidentsObj.Report != nil {
		t.Errorf("Expected nil, got %v\n", incidentsObj)
	}

	// success case
	handler = func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"Name":"ServiceNowQuery","Report":[{"number":"INC1234"}]}`)
	}

	w = httptest.NewRecorder()
	handler(w, req)

	resp = w.Result()
	err = incidentsObj.ParseBody(resp)
	if err != nil {
		t.Errorf("Expected nil, got error %v\n", err)
	}
	if incidentsObj == nil {
		t.Errorf("Expected data, got nil")
	}

}

func TestMergeIncs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := make(map[string]int)
	m["High"] = 1

	ch := make(chan map[string]int)

	// send two entries to channel ch
	go func() {
		ch <- m
		ch <- m
		close(ch)
	}()

	out := make(chan map[string]int)

	// MergeIncs should consume from ch channel and send merged output to out channel
	go func() {
		defer close(out)
		incidents.MergeIncs(ctx, ch, out)
	}()

	in := 0
	for n := range out {
		if v, ok := n["High"]; ok {
			in++
			if v != 2 {
				t.Errorf("Expected 2, got %v\n", v)
			}
		} else {
			t.Errorf("Expected `High` key, but not found")
		}
	}

	if in != 1 {
		t.Errorf("Expected 1, got %v\n", in)
	}

}

func benchmarkMergeIncs(i int, b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := make(map[string]int)
	m["High"] = 1
	m["Low"] = 1

	for n := 0; n < b.N; n++ {
		ch := make(chan map[string]int)
		go func() {
			for i := 0; i < 50; i++ {
				ch <- m
			}
			close(ch)
		}()
		out := make(chan map[string]int)
		go func() {
			defer close(out)
			incidents.MergeIncs(ctx, ch, out)
		}()
		for range out {
		}
	}
}

func BenchmarkMergeIncs10(b *testing.B)     { benchmarkMergeIncs(10, b) }
func BenchmarkMergeIncs100(b *testing.B)    { benchmarkMergeIncs(100, b) }
func BenchmarkMergeIncs500(b *testing.B)    { benchmarkMergeIncs(500, b) }
func BenchmarkMergeIncs1000(b *testing.B)   { benchmarkMergeIncs(1000, b) }
func BenchmarkMergeIncs100000(b *testing.B) { benchmarkMergeIncs(100000, b) }

// This results does not show much improvement when more goroutines and created
// TODO: need further analysis
func BenchmarkMergeAll(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := make(map[string]int)
	m["High"] = 1
	m["Low"] = 1
	for num := 1; num < 10; num += 1 {
		b.Run(fmt.Sprintf("%d", num), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				ch := make(chan map[string]int)
				go func() {
					for i := 0; i < 5; i++ {
						ch <- m
						//time.Sleep(time.Duration(rand.Intn(1000)) * time.Nanosecond)
					}
					close(ch)
				}()
				out := incidents.MergeAll(ctx, ch, num)
				for range out {
				}
			}
		})
	}

}

func TestGenerateAggReportPriority(t *testing.T) {
	// given incidents object, it should return aggregated report
	obj := []incidents.Incident{{"a", "b", "c", "d", "High", "f"},
		{"b", "b", "c", "d", "High", "f"}}
	sum, err := incidentsObj.GenerateAggReportPriority(obj)
	if err != nil {
		t.Errorf("Expected nil, got %v\n", err)
	}
	in := 0
	for _, elem := range *sum {
		in++
		if elem.Sum != 2 {
			t.Errorf("Expected 2, got %v\n", elem.Sum)
		}
	}

	if in != 1 {
		t.Errorf("Expected 1, got %v\n", in)
	}

}

func TestWalkIncs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj := []incidents.Incident{{"a", "b", "c", "d", "High", "f"}}

	// given incidents object, it should send the required map to output channel
	out, err := incidents.WalkIncs(ctx, obj)
	if err != nil {
		t.Errorf("Expected nil, got %v\n", err)
	}

	for n := range out {
		if v, ok := n["High"]; ok {
			if v != 1 {
				t.Errorf("Expected 1, got %v\n", v)
			}
		} else {
			t.Errorf("Expected `High` key, but not found")
		}
	}
}
