package utils

import (
	"os"
	"time"
)

func InvokeFunctionWithInterval(duration time.Duration, functionToInvoke func()) {
	ticker := time.NewTicker(duration)
	for {
		<-ticker.C
		functionToInvoke()
	}
}

func IsFileExist(fileName string) (bool, error) {
	_, err := os.Stat(fileName)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}
