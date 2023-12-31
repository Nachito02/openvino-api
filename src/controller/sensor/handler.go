package sensor

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"

	customHTTP "github.com/openvino/openvino-api/src/http"
	"github.com/openvino/openvino-api/src/model"
	"github.com/openvino/openvino-api/src/repository"
)

type QueryData struct {
	Harvest   string `json:"year"`
	Month     string `json:"month"`
	Day       string `json:"day"`
	WinerieID string `json:"winerie_id"`
}

func SaveSensorRecords(w http.ResponseWriter, r *http.Request) {
	hash := sha256.Sum256([]byte(r.Header.Get("Authorization")))
	secretHash := hex.EncodeToString(hash[:])

	var winery model.Winerie
	repository.DB.Table("wineries").Where("secret = ?", secretHash).First(&winery)

	if secretHash != winery.Secret {
		customHTTP.NewErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var sensorData model.SensorRecord
	err := customHTTP.DecodeJSONBody(w, r, &sensorData, map[string][]string{
		"timestamp": {"required", "date"},
	})
	if err != nil {
		return
	}
	sensorData.Winerie = winery
	repository.DB.Table("sensor_records").Create(&sensorData)
	customHTTP.ResponseJSON(w, sensorData)
}

func GetSensorRecords(w http.ResponseWriter, r *http.Request) {

	var query = "min(timestamp) as timestamp, sensor_id," +
		"avg(humidity2) as humidity2, avg(humidity1) as humidity1," +
		"avg(humidity05) as humidity05, avg(humidity005) as humidity005," +
		"max(wind_velocity) as wind_velocity, max(wind_gust) as wind_gust," +
		"avg(wind_direction) as wind_direction, avg(pressure) as pressure," +
		"max(rain) as rain, avg(temperature) as temperature," +
		"avg(humidity) as humidity, max(irradiance_ir) as irradiance_ir," +
		"max(irradiance_uv) as irradiance_uv, max(irradiance_vi) as irradiance_vi"

	var params = QueryData{}
	params.Harvest = r.URL.Query().Get("year")
	params.Month = r.URL.Query().Get("month")
	params.Day = r.URL.Query().Get("day")
	params.WinerieID = r.URL.Query().Get("winerie_id")

	log.Println(params)

	records := []model.SensorRecord{}
	stm := repository.DB.Debug().Select(query)
	stm2 := repository.DB.Debug()
	if params.WinerieID != "" {
		stm = stm.Where("winerie_id = ?", params.WinerieID)
		stm2 = stm2.Where("winerie_id = ?", params.WinerieID)
	}

	if params.Day == "" && params.Month == "" && params.Harvest != "" {
		stm.
			Where("EXTRACT(YEAR FROM timestamp) = ?", params.Harvest).
			Group("EXTRACT(DAY FROM timestamp), EXTRACT(MONTH FROM timestamp), sensor_id").
			Find(&records)
	} else if params.Day == "" && params.Month != "" && params.Harvest != "" {
		stm.
			Where("EXTRACT(MONTH FROM timestamp) = ? AND EXTRACT(YEAR FROM timestamp) = ?", params.Month, params.Harvest).
			Group("EXTRACT(DAY FROM timestamp), sensor_id").
			Find(&records)
	} else if params.Day != "" && params.Month != "" && params.Harvest != "" {
		stm.
			Where("EXTRACT(DAY FROM timestamp) = ? AND EXTRACT(MONTH FROM timestamp) = ? AND EXTRACT(YEAR FROM timestamp) = ?", params.Day, params.Month, params.Harvest).
			Group("EXTRACT(HOUR FROM timestamp), sensor_id").
			Find(&records)
	} else {
		sensordataCs := model.SensorRecord{}
		sensordataPv := model.SensorRecord{}
		sensordataMo := model.SensorRecord{}
		sensordataMe := model.SensorRecord{}
		stm2.Where("sensor_id = ?", "petit-verdot").Order("timestamp desc").Limit(1).Find(&sensordataPv)
		stm2.Where("sensor_id = ?", "cabernet-sauvignon").Order("timestamp desc").Limit(1).Find(&sensordataCs)
		stm2.Where("sensor_id = ?", "malbec-este").Order("timestamp desc").Limit(1).Find(&sensordataMe)
		stm2.Where("sensor_id = ?", "malbec-oeste").Order("timestamp desc").Limit(1).Find(&sensordataMo)
		records = []model.SensorRecord{sensordataCs, sensordataPv, sensordataMo, sensordataMe}
	}
	customHTTP.ResponseJSON(w, records)
	return
}

func GetSensorHashes(w http.ResponseWriter, r *http.Request) {

	var params = QueryData{}
	params.Harvest = r.URL.Query().Get("year")
	params.Month = r.URL.Query().Get("month")
	params.Day = r.URL.Query().Get("day")

	var hashes []string

	if params.Day == "" && params.Month == "" && params.Harvest != "" {
		repository.DB.Table("sensor_records").
			Where("EXTRACT(YEAR FROM timestamp) = ?", params.Harvest).Order("timestamp desc").
			Pluck("hash", &hashes)

	} else if params.Day == "" && params.Month != "" && params.Harvest != "" {
		repository.DB.Table("sensor_records").
			Where("EXTRACT(MONTH FROM timestamp) = ? AND EXTRACT(YEAR FROM timestamp) = ?", params.Month, params.Harvest).
			Order("timestamp desc").
			Pluck("hash", &hashes)
	} else if params.Day != "" && params.Month != "" && params.Harvest != "" {
		repository.DB.Table("sensor_records").
			Where("EXTRACT(DAY FROM timestamp) = ? AND EXTRACT(MONTH FROM timestamp) = ? AND EXTRACT(YEAR FROM timestamp) = ?", params.Day, params.Month, params.Harvest).
			Order("timestamp desc").
			Pluck("hash", &hashes)
	} else {
		var sensordataCs string
		var sensordataPv string
		var sensordataMo string
		var sensordataMe string
		repository.DB.Select("hash").Where("sensor_id = ?", "petit-verdot").Order("timestamp desc").Limit(1).First(&sensordataPv)
		repository.DB.Where("sensor_id = ?", "cabernet-sauvignon").Order("timestamp desc").Limit(1).First(&sensordataCs)
		repository.DB.Where("sensor_id = ?", "malbec-este").Order("timestamp desc").Limit(1).First(&sensordataMe)
		repository.DB.Where("sensor_id = ?", "malbec-oeste").Order("timestamp desc").Limit(1).First(&sensordataMo)
		hashes = []string{sensordataCs, sensordataPv, sensordataMo, sensordataMe}
	}
	customHTTP.ResponseJSON(w, hashes)
	return
}
