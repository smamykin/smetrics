package storage

import (
	"encoding/json"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"io/ioutil"
	"os"
)

func newFsPersister(fileName string) (*fsPersister, error) {

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return &fsPersister{}, err
	}
	return &fsPersister{
		file: file,
	}, nil
}

type fsPersister struct {
	file *os.File
}

func (f *fsPersister) flush(memStorage *MemStorage) (err error) {
	dump := memStorageDump{
		GaugeStore:   memStorage.GaugeStore(),
		CounterStore: memStorage.CounterStore(),
	}

	data, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return err
	}

	err = f.file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.file.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = f.file.Write(data)

	return err
}

func (f *fsPersister) restore(memStorage *MemStorage) (err error) {
	data, err := ioutil.ReadAll(f.file)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}

	dump := &memStorageDump{}
	err = json.Unmarshal(data, dump)
	if err != nil {
		return err
	}

	memStorage.gaugeStore = dump.GaugeStore
	memStorage.counterStore = dump.CounterStore

	return nil
}

type memStorageDump struct {
	GaugeStore   map[string]handlers.GaugeMetric
	CounterStore map[string]handlers.CounterMetric
}
