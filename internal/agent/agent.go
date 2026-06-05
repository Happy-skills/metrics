package agent

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"gopkg.in/h2non/gentleman.v2"

	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/Happy-skills/metrics/internal/repository"
)

var mapRuntimeMetrics = []string{"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc"}

func Run() {
	repository.AgentStorage = repository.NewMemStorage()

	go poolGetting(2)
	poolSending(10)
}

func poolGetting(pollInterval int) {
	for {
		time.Sleep(time.Duration(pollInterval) * time.Second)
		if err := getMetrics(); err != nil {
			fmt.Printf("getMetrics error: %s", err.Error())
		}
	}
}

func poolSending(pollInterval int) {
	for {
		time.Sleep(time.Duration(pollInterval) * time.Second)
		if err := sendMetrics("http://localhost:8080"); err != nil {
			fmt.Printf("sendMetrics error: %s", err.Error())
		}
	}
}

func getMetrics() error {
	var sVal string

	err := repository.AgentStorage.SetValue(models.Counter, "PollCount", "1")
	if err != nil {
		return err
	}
	err = repository.AgentStorage.SetValue(models.Gauge, "RandomValue", strconv.FormatFloat(rand.Float64(), 'f', -1, 64))
	if err != nil {
		return err
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mValues := reflect.ValueOf(m)
	for i := 0; i < mValues.NumField(); i++ {
		vName := mValues.Type().Field(i).Name
		for j := 0; j < len(mapRuntimeMetrics); j++ {
			if vName == mapRuntimeMetrics[j] {
				switch reflect.TypeOf(mValues.Field(i).Interface()).Kind() {
				case reflect.Uint64:
					sVal = strconv.FormatUint(mValues.Field(i).Interface().(uint64), 10)
				case reflect.Uint32:
					sVal = strconv.FormatUint(uint64(mValues.Field(i).Interface().(uint32)), 10)
				case reflect.Float64:
					sVal = strconv.FormatFloat(mValues.Field(i).Interface().(float64), 'f', -1, 64)
				}

				err = repository.AgentStorage.SetValue(models.Gauge, mValues.Type().Field(i).Name, sVal)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func sendMetrics(serverUrl string) error {
	var sVal string
	mStorage := repository.AgentStorage.GetValues()
	client := gentleman.New()
	for _, v := range mStorage {
		vType := v.MType
		switch vType {
		case models.Counter:
			sVal = strconv.FormatInt(*v.Delta, 10)
		case models.Gauge:
			sVal = strconv.FormatFloat(*v.Value, 'f', -1, 64)
		}
		//http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		url := fmt.Sprintf("%s/update/%s/%s/%s", serverUrl, v.MType, v.ID, sVal)
		req := client.Request()
		req.Method(http.MethodPost)
		req.URL(url)
		req.SetHeader("Content-Type", "text/plain")
		response, err := req.Send()
		if err != nil {
			return err
		}
		if !response.Ok {
			return errors.New(response.String())
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				panic(err)
			}
		}(response.RawResponse.Body)
	}
	return nil
}
