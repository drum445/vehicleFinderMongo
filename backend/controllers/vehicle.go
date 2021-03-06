package controllers

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/drum445/vehicleFinder/backend/models"
	"github.com/drum445/vehicleFinder/backend/repos"
	"github.com/gorilla/mux"
)

type response struct {
	Count    int             `json:"count"`
	Vehicles models.Vehicles `json:"vehicles,omitempty"`
}

func GetVehicles(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// default the page number to 1. If the page has been passed through as a URL
	// param, first check if it is an int, if it is use it, else reuturn 400
	page := 1
	var err error

	if req.URL.Query().Get("page") != "" {
		page, err = strconv.Atoi(req.URL.Query().Get("page"))
		if err != nil {
			http.Error(w, "page must be an int", 400)
			return
		}
	}

	// map[string]string of all our expected params
	m := map[string]string{
		"make":        req.URL.Query().Get("make"),
		"short_model": req.URL.Query().Get("shortModel"),
		"long_model":  req.URL.Query().Get("longModel"),
		"trim":        req.URL.Query().Get("trim"),
		"derivative":  req.URL.Query().Get("derivative"),
		"free":        req.URL.Query().Get("free"),
		"available":   "Y",
	}

	vr := repos.NewVehicleRepo()
	defer vr.Close()
	count, vehicles := vr.GetVehicles(page, m)

	// create our response object and encode to json
	var resp response
	resp.Count = count
	resp.Vehicles = vehicles
	json.NewEncoder(w).Encode(resp)
}

func GetVehicleByID(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(req)
	vehicleID, err := strconv.Atoi(params["vehicleID"])
	if err != nil {
		http.Error(w, "vehicle ID must be an int", 400)
		return
	}

	vr := repos.NewVehicleRepo()
	defer vr.Close()
	vehicle, found := vr.GetVehicle(vehicleID)

	if !found {
		http.Error(w, "vehicle ID not found", 400)
		return
	}

	vehicle.Image = repos.GetImage(vehicle.ID)
	json.NewEncoder(w).Encode(vehicle)
}

func PostVehicles(w http.ResponseWriter, req *http.Request) {
	csvFile, err := os.Open("Vehicles.csv")

	defer csvFile.Close()

	if err != nil {
		panic(err)
	}

	vr := repos.NewVehicleRepo()
	defer vr.Close()

	// load file then skip Header
	reader := csv.NewReader(bufio.NewReader(csvFile))
	reader.Read()

	// loop through each record create a vehicle object and import
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		var vehicle models.Vehicle
		vehicle.ID, _ = strconv.Atoi(record[0])
		vehicle.Make = record[1]
		vehicle.ShortModel = record[2]
		vehicle.LongModel = record[3]
		vehicle.Trim = record[4]
		vehicle.Derivative = record[5]
		vehicle.Introduced = record[6]
		vehicle.Discontinued = record[7]
		vehicle.Available = record[8]

		vr.InsertVehicle(vehicle)
	}

	fmt.Fprint(w, "Finished importing")

}
