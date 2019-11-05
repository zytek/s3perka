package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

func gb(size int64) string {
	s := fmt.Sprintf("%d GB", size/1024/1024/1024)
	return s
}

type jobStats struct {
	sync.RWMutex
	Size int64
	Num  int
}

func (js *jobStats) GetSize() int64 {
	js.RLock()
	defer js.RUnlock()
	return js.Size
}

func (js *jobStats) Add(size int64) {
	js.Lock()
	defer js.Unlock()
	js.Num++
	js.Size += size
}

func (js *jobStats) GetNum() int {
	js.RLock()
	defer js.RUnlock()
	return js.Num
}

type job struct {
	source        *Bucket
	destination   *Bucket
	copyList      []string
	copyTotalSize int64
	copyChan      chan string
	stats         jobStats
}

func (j *job) copyObject(key string) {
	tmpfile, err := ioutil.TempFile("", "lopata_tmp")
	defer os.Remove(tmpfile.Name())
	if err != nil {
		log.Fatal("Failed to create tmp file: ", err)
	}
	_, err = j.source.DownloadObject(key, tmpfile)
	if err != nil {
		log.Fatal("Failed to download object ", key, ": ", err)
	}
	destKey := path.Join(j.destination.Prefix, key)
	err = j.destination.UploadObject(destKey, tmpfile)
	if err != nil {
		log.Fatal("Failed to upload object ", destKey, ": ", err)
	}

	j.stats.Add(*j.source.Objects[key].Size)
}

func (j *job) runConsumers(wg *sync.WaitGroup) {
	for i := 0; i < cap(j.copyChan); i++ {
		go func() {
			wg.Add(1)
			for key := range j.copyChan {
				j.copyObject(key)
			}
			wg.Done()
		}()
	}
}
func (j *job) copy() {
	var wg = sync.WaitGroup{}
	j.runConsumers(&wg)
	for _, v := range j.copyList {
		j.copyChan <- v
	}
	close(j.copyChan)
	wg.Wait()
}

func (j *job) prepare() {
	j.collectKeys()

	// copy all files that are either nonexistent
	// or whose size differs
	for k, o := range j.source.Objects {
		destKey := path.Join(j.destination.Prefix, k)
		dest, found := j.destination.Objects[destKey]
		if !found || *o.Size != *dest.Size {
			j.copyList = append(j.copyList, k)
			j.copyTotalSize += *o.Size
		}
	}
}

func (j *job) collectKeys() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := j.source.CollectKeys()
		if err != nil {
			log.Fatal("Failed to list source bucket objects: ", err)
		}
		log.Println(j.name(), "source: found", j.source.TotalCount, "objects (", gb(j.source.TotalSize), ")")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := j.destination.CollectKeys()
		if err != nil {
			log.Fatal("Failed to list destination bucket objects: ", err)
		}
		log.Println(j.name(), "destination: found", j.destination.TotalCount, "objects (", gb(j.destination.TotalSize), ")")
	}()
	wg.Wait()
}

func (j *job) name() string {
	return fmt.Sprintf("[%s->%s]", j.source.Name, j.destination.Name)
}

func (j *job) status() {
	log.Println(j.name(), "status: copied", j.stats.GetNum(), "objects (", gb(j.stats.GetSize()), ")")
}

func (j *job) Start() {
	j.prepare()
	log.Println(j.name(), "starting, must copy", len(j.copyList), "objects (", gb(j.copyTotalSize), ")")
	go func() {
		for {
			time.Sleep(time.Second)
			j.status()
		}
	}()
	j.copy()
}
