package storage

import (
	"github.com/smamykin/smetrics/internal/server/handlers"
	"github.com/smamykin/smetrics/internal/utils"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const expectedJSON = `{
  "GaugeStore": {
    "metric_name3": {
      "Value": 33.44,
      "Name": "metric_name3"
    }
  },
  "CounterStore": {
    "metric_name1": {
      "Value": 11,
      "Name": "metric_name1"
    }
  }
}`

const expectedJSONAfterUpdate = `{
  "GaugeStore": {
    "metric_name3": {
      "Value": 33.44,
      "Name": "metric_name3"
    },
    "metric_name4": {
      "Value": 55.66,
      "Name": "metric_name4"
    }
  },
  "CounterStore": {
    "metric_name1": {
      "Value": 11,
      "Name": "metric_name1"
    },
    "metric_name2": {
      "Value": 22,
      "Name": "metric_name2"
    }
  }
}`

func TestFsPersister_Flush(t *testing.T) {
	fileName := "/tmp/test_fs_persister_flush.json"
	isFileExist, err := utils.IsFileExist(fileName)
	if isFileExist {
		os.Remove(fileName)
	}
	check(t, err)

	// set there a memory store
	memStorage := NewMemStorageDefault()
	// add to memory store some metrics
	memStorage.UpsertCounter(handlers.CounterMetric{Value: 11, Name: "metric_name1"})
	memStorage.UpsertGauge(handlers.GaugeMetric{Value: 33.44, Name: "metric_name3"})

	// create instance

	fsPersister := FsPersister{fileName}
	// invoke  flush
	err = fsPersister.Flush(memStorage)
	check(t, err)
	assertResultFile(t, expectedJSON, fileName)

	// upsert something to the storage again
	memStorage.UpsertCounter(handlers.CounterMetric{Value: 22, Name: "metric_name2"})
	memStorage.UpsertGauge(handlers.GaugeMetric{Value: 55.66, Name: "metric_name4"})

	// check the file again
	err = fsPersister.Flush(memStorage)
	check(t, err)
	assertResultFile(t, expectedJSONAfterUpdate, fileName)
}

func TestFsPersister_Restore(t *testing.T) {
	fileName := "/tmp/test_fs_persister_restore.json"
	isFileExist, err := utils.IsFileExist(fileName)
	if isFileExist {
		err = os.Remove(fileName)
	}
	check(t, err)

	memStorage := NewMemStorageDefault()
	fsPersister := FsPersister{fileName}
	err = fsPersister.Restore(memStorage)

	require.NotNil(t, err)
	require.Equal(t, NewMemStorageDefault(), memStorage)

	err = os.WriteFile(fileName, []byte(expectedJSON), 0664)
	check(t, err)
	fsPersister.Restore(memStorage)

	expected := NewMemStorageDefault()
	expected.UpsertCounter(handlers.CounterMetric{Value: 11, Name: "metric_name1"})
	expected.UpsertGauge(handlers.GaugeMetric{Value: 33.44, Name: "metric_name3"})
	require.Equal(t, expected, memStorage)
}

func assertResultFile(t *testing.T, expectedJSON string, fileName string) {
	dat, err := os.ReadFile(fileName)
	check(t, err)
	require.Equal(t, expectedJSON, string(dat))
}

func check(t *testing.T, e error) {
	if e != nil {
		t.Fatal(e)
	}
}
