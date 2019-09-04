/*
Incidents package defines the incidents structure and methods.
It uses servicenow format for incidents.
*/
package incidents

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
)

const (
	NumGoRoutines = 10 // maximum number of go routines to fan out
)

// Incidents Json structure
type Incidents struct {
	Name   string     `json:"Name"`
	Report []Incident `json:"Report"`
}

// Individual inc structure
type Incident struct {
	Number      string `json:"number"`
	AssignedTo  string `json:"assigned_to"`
	Description string `json:"description"`
	State       string `json:"state"`
	Priority    string `json:"priority"`
	Severity    string `json:"severity"`
}

// Aggregated report structure based on priority
type PrioritySum struct {
	Priority string
	Sum      int
}

// Initialize Incidents obj
func Init() (*Incidents, error) {
	return &Incidents{}, nil
}

// Parse the JSON response and encode it into Incidents structure
func (incidents *Incidents) ParseBody(res *http.Response) error {
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return readErr
	}
	defer res.Body.Close()

	// encode the body into json with given incidents struct

	jsonErr := json.Unmarshal(body, &incidents)
	if jsonErr != nil {
		return jsonErr
	}
	return nil
}

// walkIncs will walk through slice of incidents and sends required priority details
// to outbound channel. Once slice values are exhausted, close the output channel
// If done signal received, return early
func WalkIncs(ctx context.Context, report []Incident) (chan map[string]int, error) {
	out := make(chan map[string]int)
	go func() {
		defer close(out)
		for _, obj := range report {
			m := make(map[string]int)
			m[obj.Priority] = 1
			select {
			case out <- m:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func MergeAll(ctx context.Context, incs chan map[string]int, num int) chan map[string]int {
	//runtime.GOMAXPROCS(12)
	// Fan out the channel `incs` to bounded go routines. This will merge the values and
	// send to single output channel
	// This provides first level of merge on the data
	c := make(chan map[string]int)
	// use sync.WaitGroup for synchronization
	var wg sync.WaitGroup
	// NumGoRoutines - controls number of goroutines that can be spawned
	wg.Add(num)

	// spawn NumGoRoutines times mergeIncs()
	for i := 0; i < num; i++ {
		go func() {
			MergeIncs(ctx, incs, c)
			wg.Done()
		}()
	}

	// wait for all goroutines to end before closing channel c
	go func() {
		wg.Wait()
		close(c)
	}()

	// final merge - call mergeIncs one more time for final merge
	// make sure to close the channel
	final := make(chan map[string]int, 1)
	go func() {
		defer close(final)
		MergeIncs(ctx, c, final)
	}()
	return final
}

// Generate aggregated report based on priority
// Send all inc details into one channel
// Fan out that channel to bounded go routines. This will merge the values and
// send to single output channel
// We can have any levels of merging depending on load
func (incidents *Incidents) GenerateAggReportPriority(report []Incident) (sum *[]PrioritySum, err error) {

	// create context with cancel to inform goroutines to exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send all inc details into one channel incs
	incs, errc := WalkIncs(ctx, report)
	if errc != nil {
		return nil, errc
	}

	final := MergeAll(ctx, incs, NumGoRoutines)

	// read the final channel and create []PrioritySum struct
	var sumObj []PrioritySum
	for obj := range final {
		for k, v := range obj {
			sumObj = append(sumObj, PrioritySum{k, v})
		}
	}
	return &sumObj, nil
}

// Merges the inputs based on priority
// Since input and outbound channels are same, we can compose this any number of times
// No need to close out channel because it is used by multiple go routines. Main has to close it
// For the same reason, we do not need context or done channels
func MergeIncs(ctx context.Context, incs chan map[string]int, out chan map[string]int) {
	sev := make(map[string]int)
	for obj := range incs {
		for k, v := range obj {
			// aggregate same key objects
			if _, ok := sev[k]; ok {
				sev[k] += v
			} else {
				sev[k] = v
			}
		}
	}
	// send aggregated value
	select {
	case out <- sev:
	case <-ctx.Done():
	}
	return
}
