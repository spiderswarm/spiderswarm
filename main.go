package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/davecgh/go-spew/spew"
)

type DataPipe struct {
	Done  bool
	Queue []interface{}
}

func NewDataPipe() *DataPipe {
	return &DataPipe{false, []interface{}{}}
}

func (dp *DataPipe) Add(x interface{}) {
	dp.Queue = append(dp.Queue, x)
}

func (dp *DataPipe) Remove() interface{} {
	if len(dp.Queue) == 0 {
		return nil
	}

	lastIdx := len(dp.Queue) - 1
	x := dp.Queue[lastIdx]
	dp.Queue = dp.Queue[:lastIdx]

	return x
}

type Action interface {
	Run() error
	AddInput(name string, dataPipe DataPipe)
	AddOutput(name string, dataPipe DataPipe)
}

type AbstractAction struct {
	Action
	Inputs             map[string]*DataPipe
	Outputs            map[string]*DataPipe
	CanFail            bool
	ExpectMany         bool
	AllowedInputNames  []string
	AllowedOutputNames []string
}

func (a *AbstractAction) AddInput(name string, dataPipe *DataPipe) error {
	for _, n := range a.AllowedInputNames {
		if n == name {
			a.Inputs[name] = dataPipe
			return nil
		}
	}

	return errors.New("input name not in AllowedInputNames")
}

func (a *AbstractAction) AddOutput(name string, dataPipe *DataPipe) error {
	for _, n := range a.AllowedOutputNames {
		if n == name {
			a.Outputs[name] = dataPipe
			return nil
		}
	}

	return errors.New("input name not in AllowedOutputNames")
}

func (a *AbstractAction) Run() error {
	// To be implemented by concrete actions.
	return nil
}

const HTTPActionInputURLParams = "HTTPActionInputURLParams"
const HTTPActionInputHeaders = "HTTPActionInputHeaders"
const HTTPActionInputCookies = "HTTPActionInputCookies"

const HTTPActionOutputBody = "HTTPActionOutputBody"
const HTTPActionOutputHeaders = "HTTPActionOutputHeaders"
const HTTPActionOutputStatusCode = "HTTPActionOutputStatusCode"

type HTTPAction struct {
	AbstractAction
	BaseURL string
	Method  string
}

func NewHTTPAction(baseURL string, method string, canFail bool) *HTTPAction {
	return &HTTPAction{
		AbstractAction: AbstractAction{
			CanFail:    canFail,
			ExpectMany: false,
			AllowedInputNames: []string{
				HTTPActionInputURLParams,
				HTTPActionInputHeaders,
				HTTPActionInputCookies,
			},
			AllowedOutputNames: []string{
				HTTPActionOutputBody,
				HTTPActionOutputHeaders,
				HTTPActionOutputStatusCode,
			},
			Inputs:  map[string]*DataPipe{},
			Outputs: map[string]*DataPipe{},
		},
		BaseURL: baseURL,
		Method:  method,
	}
}

func (ha *HTTPAction) Run() error {
	request, err := http.NewRequest(ha.Method, ha.BaseURL, nil)
	if err != nil {
		return err
	}

	q := request.URL.Query()

	if ha.Inputs[HTTPActionInputURLParams] != nil {
		for {
			urlParams, ok := ha.Inputs[HTTPActionInputURLParams].Remove().(map[string][]string)
			if !ok {
				break
			}

			for key, values := range urlParams {
				for _, value := range values {
					q.Add(key, value)
				}
			}
		}
	}

	if ha.Inputs[HTTPActionInputHeaders] != nil {
		for {
			headers, ok := ha.Inputs[HTTPActionInputHeaders].Remove().(map[string][]string)

			if !ok {
				break
			}

			for key, values := range headers {
				for _, value := range values {
					request.Header.Add(key, value)
				}
			}

		}
	}

	if ha.Inputs[HTTPActionInputCookies] != nil {
		for {
			cookies, ok := ha.Inputs[HTTPActionInputCookies].Remove().(map[string]string)

			if !ok {
				break
			}

			for key, value := range cookies {
				c := &http.Cookie{Name: key, Value: value}
				request.AddCookie(c)
			}

		}
	}

	request.URL.RawQuery = q.Encode()

	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	if ha.Outputs[HTTPActionOutputBody] != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			ha.Outputs[HTTPActionOutputBody].Add(body)
		}
	}

	if ha.Outputs[HTTPActionOutputHeaders] != nil {
		headers := resp.Header
		ha.Outputs[HTTPActionOutputHeaders].Add(headers)
	}

	if ha.Outputs[HTTPActionOutputStatusCode] != nil {
		statusCode := resp.StatusCode
		ha.Outputs[HTTPActionOutputStatusCode].Add(statusCode)
	}

	return nil
}

type Task struct {
	Inputs  map[string]*DataPipe
	Outputs map[string]*DataPipe
}

type Workflow struct {
	Name    string
	Version string
	Tasks   []Task
}

func main() {
	fmt.Println("SpiderSwarm")
	httpAction := NewHTTPAction("https://ifconfig.me/", "GET", true)

	headers := map[string][]string{
		"User-Agent": []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"},
	}

	headersIn := NewDataPipe()

	headersIn.Add(headers)

	httpAction.AddInput(HTTPActionInputHeaders, headersIn)

	bodyOut := NewDataPipe()
	httpAction.AddOutput(HTTPActionOutputBody, bodyOut)

	headersOut := NewDataPipe()
	httpAction.AddOutput(HTTPActionOutputHeaders, headersOut)

	statusCodeOut := NewDataPipe()
	httpAction.AddOutput(HTTPActionOutputStatusCode, statusCodeOut)

	spew.Dump(httpAction)

	err := httpAction.Run()
	if err != nil {
		fmt.Println(err)
	}

	spew.Dump(httpAction)
}
