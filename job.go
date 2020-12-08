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

func bytes(b int64) string {
    const unit = 1024
    if b < unit {
        return fmt.Sprintf("%d B", b)
    }
    div, exp := int64(unit), 0
    for n := b / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %ciB",
        float64(b)/float64(div), "KMGTPE"[exp])
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
		wg.Add(1)
		go func() {
			for k := range j.copyChan {
				j.copyObject(k)
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
  j.status()
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
		log.Println(j.name(), "source: found", j.source.TotalCount, "objects (", bytes(j.source.TotalSize), ")")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := j.destination.CollectKeys()
		if err != nil {
			log.Fatal("Failed to list destination bucket objects: ", err)
		}
		log.Println(j.name(), "destination: found", j.destination.TotalCount, "objects (", bytes(j.destination.TotalSize), ")")
	}()
	wg.Wait()
}

func (j *job) name() string {
	return fmt.Sprintf("[%s->%s]", j.source.Name, j.destination.Name)
}

func (j *job) status() {
	log.Println(j.name(), "status update: copied", j.stats.GetNum(), "objects (", bytes(j.stats.GetSize()), ")")
}

func (j *job) Start() {
	j.prepare()
	log.Println(j.name(), "starting, must copy", len(j.copyList), "new objects (", bytes(j.copyTotalSize), ")")
	go func() {
		for {
			time.Sleep(time.Second)
			j.status()
		}
	}()
	j.copy()
}
