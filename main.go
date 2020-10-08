package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const sesionPortalServletURL = "http://pocae.tstgo.cl/PortalCAE-WAR-MODULE/SesionPortalServlet"
const comercialesPortalServletURL = "http://pocae.tstgo.cl/PortalCAE-WAR-MODULE/ComercialesPortalServlet"

func main() {
	http.HandleFunc("/api/bipcard/info/", handleBipCardBalance)
	http.HandleFunc("/api/bipcard/movements/", handleBipCardMovements)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

func handleBipCardBalance(w http.ResponseWriter, r *http.Request) {
	cardNumber := r.URL.Query().Get("card_number")

	if cardNumber == "" {
		http.Error(w, "card_number param is required", http.StatusBadRequest)
		return
	}

	bodyBalanceRequest := fmt.Sprintf("accion=6&NumDistribuidor=99&NomUsuario=usuInternet&NomHost=AFT&NomDominio=aft.cl&Trx=&RutUsuario=0&NumTarjeta=%s&bloqueable=", cardNumber)

	body, err := getBytesPostRequest(sesionPortalServletURL, bodyBalanceRequest)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	re1, err := regexp.Compile(`<td.*class="verdanabold-ckc">(.*)</td>`)
	result := re1.FindAllStringSubmatch(string(body), -1)

	if len(result) != 8 {
		http.Error(w, "scrapping error: unexpected changes in webpage ", http.StatusInternalServerError)
		return
	}

	cardInfo := cardInfo{
		CardNumber:      result[1][1],
		ContractStatus:  result[3][1],
		CardBalance:     result[5][1],
		CardBalanceDate: result[7][1],
	}

	js, err := json.Marshal(cardInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func handleBipCardMovements(w http.ResponseWriter, r *http.Request) {
	cardNumber := r.URL.Query().Get("card_number")

	if cardNumber == "" {
		http.Error(w, "card_number param is required", http.StatusBadRequest)
		return
	}

	bodyBalanceRequest := fmt.Sprintf("accion=6&NumDistribuidor=99&NomUsuario=usuInternet&NomHost=AFT&NomDominio=aft.cl&Trx=&RutUsuario=0&NumTarjeta=%s&bloqueable=", cardNumber)

	body, err := getBytesPostRequest(sesionPortalServletURL, bodyBalanceRequest)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	re1, err := regexp.Compile(`<input\s*type="hidden"\s*id="KSI"\s*name="KSI"\s*value="(.*)">`)
	result := re1.FindAllStringSubmatch(string(body), -1)

	if len(result) == 0 {
		http.Error(w, "scrapping error: unexpected changes in webpage ", http.StatusInternalServerError)
		return
	}

	t := time.Now().UTC()
	fechaInicioMovimientos := fmt.Sprintf("%d%02d%02d", t.Year(), t.Month(), t.Day())
	fechalogeo := fmt.Sprintf("%d%02d%02d%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	ksi := result[0][1]

	bodyMovementsRequest := fmt.Sprintf("accion=1&KSI=%s&DiasMov=90&FechaInicioMovimientos=%s&itemms=3000&item=2&fechalogeo=%s&cboSeleccion=90", ksi, fechaInicioMovimientos, fechalogeo)

	println(bodyMovementsRequest)

	body, err = getBytesPostRequest(comercialesPortalServletURL, bodyMovementsRequest)

	re2, err := regexp.Compile(`<tr id="fila_[\s\S]*?<\/tr>`)
	resultDetails := re2.FindAllStringSubmatch(string(body), -1)

	if len(resultDetails) == 0 {
		http.Error(w, "scrapping error: unexpected changes in webpage ", http.StatusInternalServerError)
		return
	}

	var dataResponse []cardMovement

	for _, val := range resultDetails {

		reDetail, _ := regexp.Compile(`<td[\s\S]*?>(.*)<\/td>`)
		reResult := reDetail.FindAllStringSubmatch(val[0], -1)

		movement := cardMovement{
			MovementID:   strings.Replace(reResult[1][1], "&nbsp;", "", -1),
			TypeMovement: strings.Replace(reResult[2][1], "&nbsp;", "", -1),
			DateTime:     strings.Replace(reResult[3][1], "&nbsp;", "", -1),
			Place:        strings.Replace(reResult[4][1], "&nbsp;", "", -1),
			Amount:       strings.Replace(reResult[5][1], "&nbsp;", "", -1),
			Balance:      strings.Replace(reResult[6][1], "&nbsp;", "", -1),
		}

		dataResponse = append(dataResponse, movement)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(dataResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

type cardInfo struct {
	CardNumber      string `json:"cardNumber"`
	ContractStatus  string `json:"contractStatus"`
	CardBalance     string `json:"cardBalance"`
	CardBalanceDate string `json:"cardBalanceDate"`
}

type cardMovement struct {
	MovementID   string `json:"movementId"`
	TypeMovement string `json:"typeMovement"`
	DateTime     string `json:"dateTime"`
	Place        string `json:"place"`
	Amount       string `json:"amount"`
	Balance      string `json:"balance"`
}

func getBytesPostRequest(url string, formBody string) ([]byte, error) {
	bodyBytes := []byte(formBody)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
