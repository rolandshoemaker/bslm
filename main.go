package main

import (
	"net/http"
	"sync"
)

type sth struct {
	timestamp int
	treeSize  int
}

type data struct {
	mu        sync.Mutex
	sthMap    map[string]sth
	lag       map[string]int
	frontends []string
}

type respSTH struct {
	TreeSize  int `json:"tree_size"`
	Timestamp int `json:"timestamp"`
}

func (d *data) lookup() {
	wg := new(sync.WaitGroup)
	for _, frontend := range d.frontends {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			// lookup sth from frontend
			resp, err := http.Get(fmt.Sprintf("%s/ct/v1/get-sth", f))
			if err != nil {
				// log
				return
			}
			defer close(resp.Body)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				// log
				return
			}
			var returned respSTH
			err = json.Unmarshal(body, &returned)
			if err != nil {
				// log
				return
			}
			// store in d.sthMap
			d.sthMap[f] = sth{
				timestamp: returned.Timestamp,
				treeSize:  returned.TreeSize,
			}
		}(frontend)
	}
	wg.wait()
}

func (d *data) updateLag() {
	var newestSTH *sth
	d.mu.RLock()
	defer d.mu.RUnlock()
	for _, sth := range d.sthMap {
		if newestSTH == nil || newestSTH.timestamp < sth.timestamp {
			newestSTH = &sth
		}
	}
	for frontend, sth := range d.sthMap {
		d.lag[frontend] = newestSTH.timestamp - sth.timestamp
	}
}
