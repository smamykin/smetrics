package storage

import (
	"encoding/json"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"os"
)

func newFsPersister(fileName string) *fsPersister {
	return &fsPersister{
		fileName: fileName,
	}
}

type fsPersister struct {
	fileName string
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

	return os.WriteFile(f.fileName, data, 0664)
}

func (f *fsPersister) restore(memStorage *MemStorage) (err error) {
	data, err := os.ReadFile(f.fileName)
	if err != nil {
		return err
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
