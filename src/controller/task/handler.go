package task

import (
	"net/http"
	"time"

	customHTTP "github.com/openvino/openvino-api/src/http"
	"github.com/openvino/openvino-api/src/model"
	"github.com/openvino/openvino-api/src/repository"
	"github.com/thedevsaddam/govalidator"
)

type QueryData struct {
	Harvest   string `json:"year"`
	Month     string `json:"month"`
	Day       string `json:"day"`
	PublicKey string `json:"public_key"`
	WinerieID string `json:"winerie_id"`
}

type InsertData struct {
	Hash            string     `json:"hash"`
	PublicKey       string     `json:"public_key"`
	IniTimestamp    *time.Time `json:"ini_timestamp"`
	IniClaro        string     `json:"ini_claro"`
	IniRow          uint       `json:"ini_row"`
	IniPlant        uint       `json:"ini_plant"`
	EndTimestamp    *time.Time `json:"end_timestamp"`
	EndClaro        string     `json:"end_claro"`
	EndRow          uint       `json:"end_row"`
	EndPlant        uint       `json:"end_plant"`
	CategoryId      uint       `json:"category_id"`
	TypeId          uint       `json:"task_id"`
	ToolsUsed       []uint     `json:"tools_used"`
	Chemicals       []uint     `json:"chemicals"`
	ChemicalAmounts []float32  `json:"chemicals_amount"`
	Notes           string     `json:"notes"`
	WinerieID       string     `json:"winerie_id"`
}

type ToolsData struct {
	Id uint `json:"id"`
}

func GetTasks(w http.ResponseWriter, r *http.Request) {

	var params = QueryData{}
	params.Harvest = r.URL.Query().Get("year")
	params.Month = r.URL.Query().Get("month")
	params.Day = r.URL.Query().Get("day")
	params.PublicKey = r.URL.Query().Get("public_key")
	params.WinerieID = r.URL.Query().Get("winerie_id")

	tasks := []model.Task{}

	query := repository.DB

	if params.Day == "" && params.Month == "" && params.Harvest != "" {
		query = query.Where("EXTRACT(YEAR FROM ini_timestamp) = ?", params.Harvest)
	}
	if params.Month != "" {
		query = query.Where("EXTRACT(MONTH FROM ini_timestamp) = ?", params.Month)
	}
	if params.Day != "" {
		query = query.Where("EXTRACT(DAY FROM ini_timestamp) = ?", params.Day)
	}
	if params.PublicKey != "" {
		query = query.Where("public_key = ?", params.PublicKey)
	}
	if params.WinerieID != "" {
		query = query.Where("winerie_id = ?", params.WinerieID)
	}

	query.Preload("ToolsUsed").Preload("ChemicalsUsed").Order("ini_timestamp desc").Find(&tasks)
	customHTTP.ResponseJSON(w, tasks)
	return
}

func CreateTask(w http.ResponseWriter, r *http.Request) {

	var body InsertData
	rules := govalidator.MapData{
		"hash":             []string{"required", "string"},
		"public_key":       []string{"required", "string"},
		"ini_timestamp":    []string{"required", "date"},
		"ini_claro":        []string{"required", "string"},
		"ini_row":          []string{"required", "uint"},
		"ini_plant":        []string{"required", "uint"},
		"end_timestamp":    []string{"required", "date"},
		"end_claro":        []string{"required", "string"},
		"end_row":          []string{"required", "uint"},
		"end_plant":        []string{"required", "uint"},
		"category_id":      []string{"required", "uint"},
		"task_id":          []string{"required", "uint"},
		"tools_used":       []string{"required", "[]uint"},
		"chemicals":        []string{"optional", "[]string"},
		"chemicals_amount": []string{"optional", "[]float32"},
		"notes":            []string{"optional", "string"},
		"winerie_id":       []string{"required", "int"},
	}
	err := customHTTP.DecodeJSONBody(w, r, &body, rules)
	if err != nil {
		customHTTP.NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var winerie model.Winerie
	err = repository.DB.First(&winerie, body.WinerieID).Error
	if err != nil {
		customHTTP.NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	task := model.Task{
		Hash:         body.Hash,
		PublicKey:    body.PublicKey,
		IniTimestamp: body.IniTimestamp,
		IniClaro:     body.IniClaro,
		IniRow:       body.IniRow,
		IniPlant:     body.IniPlant,
		EndTimestamp: body.EndTimestamp,
		EndClaro:     body.EndClaro,
		EndRow:       body.EndRow,
		EndPlant:     body.EndPlant,
		CategoryId:   body.CategoryId,
		TypeId:       body.TypeId,
		Notes:        body.Notes,
		WinerieID:    body.WinerieID,
	}

	for _, element := range body.ToolsUsed {
		task.ToolsUsed = append(task.ToolsUsed, model.Tools{
			Id:       element,
			TaskHash: body.Hash,
		})
	}

	for i, element := range body.Chemicals {
		task.ChemicalsUsed = append(task.ChemicalsUsed, model.Chemicals{
			Id:       element,
			Amount:   body.ChemicalAmounts[i],
			TaskHash: body.Hash,
		})
	}

	repository.DB.Create(task)

}
