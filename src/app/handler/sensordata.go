package handler

import (
	"time"
	"encoding/json"
	"net/http"
	"github.com/jinzhu/gorm"
	"github.com/openvino/openvino-api/src/app/model"
)

func GetSensorDataWrong(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	respondError(w, http.StatusBadRequest, "Malformed query")
}

func GetSensorDataDay(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	
	day := r.URL.Query().Get("day")
	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")

	sensordata := []model.SensorData{}

	db.Where("DAY(timestamp) = ? AND MONTH(timestamp) = ? AND YEAR(timestamp) = ?", day, month, year).Find(&sensordata)
	respondJSON(w, http.StatusOK, sensordata)
	
}

func GetSensorDataMonth(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	
	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")

	sensordata := []model.SensorData{}

	db.Select("min(timestamp) as timestamp, sensor_id," +
				"avg(humidity2) as humidity2, avg(humidity1) as humidity1," +
				"avg(humidity05) as humidity05, avg(humidity005) as humidity005," +
				"max(wind_velocity) as wind_velocity, max(wind_gust) as wind_gust," +
				"avg(wind_direction) as wind_direction, avg(pressure) as pressure," +
				"max(rain) as rain, avg(temperature) as temperature," +
				"avg(humidity) as humidity").Group("day(timestamp), sensor_id").Having("year(timestamp) = ? AND month(timestamp) = ?", year, month).Find(&sensordata)

	respondJSON(w, http.StatusOK, sensordata)

}

func GetSensorDataYear(db *gorm.DB, w http.ResponseWriter, r *http.Request) {

	year := r.URL.Query().Get("year")

	sensordata := []model.SensorData{}

	db.Select("min(timestamp) as timestamp, sensor_id," +
				"avg(humidity2) as humidity2, avg(humidity1) as humidity1," +
				"avg(humidity05) as humidity05, avg(humidity005) as humidity005," +
				"max(wind_velocity) as wind_velocity, max(wind_gust) as wind_gust," +
				"avg(wind_direction) as wind_direction, avg(pressure) as pressure," +
				"max(rain) as rain, avg(temperature) as temperature," +
				"avg(humidity) as humidity").Group("month(timestamp), sensor_id").Having("year(timestamp) = ?", year).Find(&sensordata)

	respondJSON(w, http.StatusOK, sensordata)

}