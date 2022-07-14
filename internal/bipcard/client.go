package bipcard

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const sesionPortalServletURL = "http://pocae.tstgo.cl/PortalCAE-WAR-MODULE/SesionPortalServlet"
const comercialesPortalServletURL = "http://pocae.tstgo.cl/PortalCAE-WAR-MODULE/ComercialesPortalServlet"

const fmtBodyBalanceRequest = "accion=6&NumDistribuidor=99&NomUsuario=usuInternet&NomHost=AFT&NomDominio=aft.cl&Trx=&RutUsuario=0&NumTarjeta=%s&bloqueable="
const fmtBodyMovementsRequest = "accion=1&KSI=%s&DiasMov=90&FechaInicioMovimientos=%s&itemms=3000&item=2&fechalogeo=%s&cboSeleccion=90"

const cardInfoDataCellRegex = `<td.*class="verdanabold-ckc">(.*)</td>`
const cardMovementsInputRegex = `<input\s*type="hidden"\s*id="KSI"\s*name="KSI"\s*value="(.*)">`
const cardMovementsRowRegex = `<tr id="fila_[\s\S]*?<\/tr>`
const cardMovementsRowDataRegex = `<td[\s\S]*?>(.*)<\/td>`

var errScrapping = errors.New("scrapping error: unexpected changes in webpage")

func NewBipCardClient() BipCardClient {
	cardInfoDataCellRegexComp, _ := regexp.Compile(cardInfoDataCellRegex)
	cardMovementsInputRegexComp, _ := regexp.Compile(cardMovementsInputRegex)
	cardMovementsRowRegexComp, _ := regexp.Compile(cardMovementsRowRegex)
	cardMovementsRowDataRegexComp, _ := regexp.Compile(cardMovementsRowDataRegex)

	return BipCardClient{
		cardInfoDataCellRegexCompiled:     cardInfoDataCellRegexComp,
		cardMovementsInputRegexCompiled:   cardMovementsInputRegexComp,
		cardMovementsRowRegexCompiled:     cardMovementsRowRegexComp,
		cardMovementsRowDataRegexCompiled: cardMovementsRowDataRegexComp,
	}
}

type BipCardClient struct {
	cardInfoDataCellRegexCompiled     *regexp.Regexp
	cardMovementsInputRegexCompiled   *regexp.Regexp
	cardMovementsRowRegexCompiled     *regexp.Regexp
	cardMovementsRowDataRegexCompiled *regexp.Regexp
}

type CardInfo struct {
	CardNumber      string    `json:"cardNumber"`
	ContractStatus  string    `json:"contractStatus"`
	CardBalance     int       `json:"cardBalance"`
	CardBalanceDate time.Time `json:"cardBalanceDate"`
}

type CardMovement struct {
	MovementID   int       `json:"movementId"`
	TypeMovement string    `json:"typeMovement"`
	DateTime     time.Time `json:"dateTime"`
	Place        string    `json:"place"`
	Amount       int       `json:"amount"`
	Balance      int       `json:"balance"`
}

func (client *BipCardClient) GetBipCardInfo(cardNumber string) (*CardInfo, error) {
	bodyBalanceRequest := fmt.Sprintf(fmtBodyBalanceRequest, cardNumber)

	body, err := getBytesPostRequest(sesionPortalServletURL, bodyBalanceRequest)

	if err != nil {
		return nil, err
	}

	result := client.cardInfoDataCellRegexCompiled.FindAllStringSubmatch(string(body), -1)

	if len(result) != 8 {
		return nil, errScrapping
	}

	strContractStatus := strings.TrimSpace(html.UnescapeString(result[3][1]))

	strCardBalance := strings.TrimSpace(html.UnescapeString(result[5][1]))
	strCardBalance = strings.Replace(strCardBalance, "$", "", -1)
	strCardBalance = strings.Replace(strCardBalance, ".", "", -1)
	intCardBalance, _ := strconv.Atoi(strCardBalance)

	strCardBalanceDate := strings.TrimSpace(html.UnescapeString(result[7][1]))
	dtCardBalanceDate, _ := time.Parse("02/01/2006 15:04", strCardBalanceDate)

	cardInfo := &CardInfo{
		CardNumber:      result[1][1],
		ContractStatus:  strContractStatus,
		CardBalance:     intCardBalance,
		CardBalanceDate: dtCardBalanceDate,
	}

	return cardInfo, nil
}

func (client *BipCardClient) GetBipCardMovements(cardNumber string) ([]CardMovement, error) {
	bodyBalanceRequest := fmt.Sprintf(fmtBodyBalanceRequest, cardNumber)

	body, err := getBytesPostRequest(sesionPortalServletURL, bodyBalanceRequest)

	if err != nil {
		return nil, err
	}

	result := client.cardMovementsInputRegexCompiled.FindAllStringSubmatch(string(body), -1)

	if len(result) == 0 {
		return nil, errScrapping
	}

	t := time.Now().UTC()
	fechaInicioMovimientos := fmt.Sprintf("%d%02d%02d", t.Year(), t.Month(), t.Day())
	fechalogeo := fmt.Sprintf("%d%02d%02d%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	ksi := result[0][1]

	bodyMovementsRequest := fmt.Sprintf(fmtBodyMovementsRequest, ksi, fechaInicioMovimientos, fechalogeo)

	body, err = getBytesPostRequest(comercialesPortalServletURL, bodyMovementsRequest)

	if err != nil {
		return nil, err
	}

	resultDetails := client.cardMovementsRowRegexCompiled.FindAllStringSubmatch(string(body), -1)

	if len(resultDetails) == 0 {
		return nil, errScrapping
	}

	var movements []CardMovement

	for _, val := range resultDetails {
		reResult := client.cardMovementsRowDataRegexCompiled.FindAllStringSubmatch(val[0], -1)

		strMovementID := strings.TrimSpace(html.UnescapeString(reResult[1][1]))
		intMovementID, _ := strconv.Atoi(strMovementID)

		strTypeMovement := strings.TrimSpace(html.UnescapeString(reResult[2][1]))

		strDateTime := strings.TrimSpace(html.UnescapeString(reResult[3][1]))
		dtDateTime, _ := time.Parse("02/01/2006 15:04", strDateTime)

		strPlace := strings.TrimSpace(html.UnescapeString(reResult[4][1]))

		strAmount := strings.Replace(strings.TrimSpace(html.UnescapeString(reResult[5][1])), ".", "", -1)

		intAmount, _ := strconv.Atoi(strAmount)

		strBalance := strings.Replace(strings.TrimSpace(html.UnescapeString(reResult[6][1])), ".", "", -1)

		intBalance, _ := strconv.Atoi(strBalance)

		movements = append(movements, CardMovement{
			MovementID:   intMovementID,
			TypeMovement: strTypeMovement,
			DateTime:     dtDateTime,
			Place:        strPlace,
			Amount:       intAmount,
			Balance:      intBalance,
		})
	}

	if err != nil {
		return nil, err
	}

	return movements, nil
}

func getBytesPostRequest(url string, formBody string) ([]byte, error) {
	bodyBytes := []byte(formBody)

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
