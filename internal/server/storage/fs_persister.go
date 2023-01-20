package storage

import (
	"encoding/json"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"os"
)

func NewFsPersister(fileName string) *FsPersister {
	return &FsPersister{
		fileName: fileName,
	}
}

type FsPersister struct {
	fileName string
}

func (f *FsPersister) Flush(memStorage *MemStorage) (err error) {
	dump := memStorageDump{
		GaugeStore:   memStorage.GaugeStore(),
		CounterStore: memStorage.CounterStore(),
	}

	data, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return
	}

	err = os.WriteFile(f.fileName, data, 0664)

	return
}

func (f *FsPersister) Restore(memStorage *MemStorage) (err error) {
	data, err := os.ReadFile(f.fileName)
	if err != nil {
		return
	}

	dump := &memStorageDump{}
	err = json.Unmarshal(data, dump)
	if err != nil {
		return
	}

	memStorage.gaugeStore = dump.GaugeStore
	memStorage.counterStore = dump.CounterStore

	return nil
}

type memStorageDump struct {
	GaugeStore   map[string]handlers.GaugeMetric
	CounterStore map[string]handlers.CounterMetric
}
