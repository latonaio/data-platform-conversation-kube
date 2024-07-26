package services

import (
	"bytes"
	apiInputReader "data-platform-conversation-kube/api-input-reader/types"
	"data-platform-conversation-kube/config"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/google/uuid"
	"github.com/latonaio/golang-logging-library-for-data-platform/logger"
	"golang.org/x/xerrors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	POST = "POST"
	GET  = "GET"
)

type RequestWrapperController struct {
	Controller   *beego.Controller
	CustomLogger *logger.Logger
}

type ResponseData struct {
	StatusCode int    `json:"statusCode"`
	Name       string `json:"name"`
	Message    string `json:"message"`
	Data       struct {
		RuntimeSessionID *string `json:"runtimeSessionId"`
	} `json:"data"`
}

type AuthenticatorResponseData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func UserRequestParams(
	requestWrapperController RequestWrapperController,
) *apiInputReader.Request {
	businessPartner, _ := requestWrapperController.Controller.GetInt("businessPartner")
	businessPartnerRole := requestWrapperController.Controller.GetString("businessPartnerRole")
	language := requestWrapperController.Controller.GetString("language")
	userId := requestWrapperController.Controller.GetString("userId")

	runtimeSessionId := uuid.New().String()
	runtimeSessionId = strings.ReplaceAll(runtimeSessionId, "-", "")

	if requestWrapperController.CustomLogger != nil {
		requestWrapperController.CustomLogger.Info(
			"RuntimeSessionID: %v",
			runtimeSessionId,
		)
	}

	return &apiInputReader.Request{
		Language:            &language,
		BusinessPartner:     &businessPartner,
		BusinessPartnerRole: &businessPartnerRole,
		UserID:              &userId,
		RuntimeSessionID:    &runtimeSessionId,
	}
}

func Request(
	aPIServiceName string,
	aPIType string,
	body io.ReadCloser,
	controller *beego.Controller,
) []byte {
	conf := config.NewConf()
	nestjsURL := conf.REQUEST.RequestURL()
	jwtToken := controller.Ctx.Input.Header("Authorization")

	method := POST
	requestUrl := fmt.Sprintf("%s/%s/%s", nestjsURL, aPIServiceName, aPIType)

	byteBody, err := ioutil.ReadAll(body)
	if err != nil {
		HandleError(
			controller,
			err,
			nil,
		)
	}

	req, err := http.NewRequest(
		method, requestUrl, ioutil.NopCloser(bytes.NewReader(byteBody)),
	)

	if err != nil {
		HandleError(
			controller,
			err,
			nil,
		)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", jwtToken)

	client := &http.Client{}

	response, err := client.Do(req)

	responseBody, err := ioutil.ReadAll(response.Body)

	err = response.Body.Close()
	if err != nil {
		HandleError(
			controller,
			err,
			nil,
		)
		return nil
	}

	if response.StatusCode != 200 && response.StatusCode != 201 {
		HandleError(
			controller,
			responseBody,
			&response.StatusCode,
		)
		return nil
	}

	return responseBody
}

func HandleError(
	controller *beego.Controller,
	message interface{},
	statusCode *int,
) {
	l := logger.NewLogger()
	ctx := controller.Ctx

	responseData := ResponseData{}

	if statusCode == nil {
		ctx.Output.SetStatus(500)
	} else {
		ctx.Output.SetStatus(*statusCode)
	}

	if msg, ok := message.([]byte); ok {
		err := json.Unmarshal(msg, &responseData)

		controller.Data["json"] = responseData
		controller.ServeJSON()

		if err != nil {
			l.Error(xerrors.Errorf("HandleError error: %w", err))
		}
	}

	if errMsg, ok := message.(error); ok {
		responseData = ResponseData{
			StatusCode: func() int {
				if statusCode != nil {
					return *statusCode
				}
				return 500
			}(),
			// todo エラーの種類をまとめておくこと
			Name:    "InternalServerError",
			Message: errMsg.Error(),
			Data: struct {
				RuntimeSessionID *string `json:"runtimeSessionId"`
			}{},
		}
	}

	controller.Data["json"] = responseData
	controller.ServeJSON()

	if statusCode != nil {
		controller.Abort(fmt.Sprintf("%d", &statusCode))
	} else {
		controller.Abort("500")
	}
}

func Respond(
	controller *beego.Controller,
	data interface{},
) {
	controller.Data["json"] = data
	controller.ServeJSON()
}
