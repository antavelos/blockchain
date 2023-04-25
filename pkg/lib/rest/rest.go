package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/antavelos/blockchain/pkg/common"
)

func handleResponse(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, common.GenericError{Msg: string(body)}
	}

	return body, err
}

func GetHttpData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

func PostHttpData(url string, data []byte) ([]byte, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

type Marshaller interface {
	Unmarshal([]byte) (any, error)
}

type ObjectMarshaller struct {
	Many bool
}

type Response struct {
	Body []byte
	Err  error
}

func (r Response) IsConnectionRefused() bool {
	if r.Err == nil {
		return false
	}
	return strings.Contains(r.Err.Error(), "connection refused")
}

type BulkResponse []Response

func (br BulkResponse) HasConnectionRefused() bool {
	connectionRefusedResponses := common.Filter(br, func(response Response) bool {
		return response.IsConnectionRefused()
	})

	return len(connectionRefusedResponses) > 0
}

func (br BulkResponse) errorResponses() BulkResponse {
	return common.Filter(br, func(r Response) bool {
		return r.Err != nil
	})
}

func (br BulkResponse) ErrorsRatio() float64 {
	return float64(len(br.errorResponses())) / float64(len(br))
}

func (br BulkResponse) ErrorStrings() []string {
	return common.Map(br.errorResponses(), func(r Response) string {
		return r.Err.Error()
	})
}

type Requester interface {
	Request() Response
}

type PostRequester struct {
	URL  string
	Body any
	M    Marshaller
}

func (r PostRequester) Request() Response {
	dataBytes, err := json.Marshal(r.Body)
	if err != nil {
		return Response{Err: err}
	}

	responseBody, err := PostHttpData(r.URL, dataBytes)
	if err != nil {

		return Response{Err: common.GenericError{Msg: r.URL, Extra: err}}
	}

	return Response{Body: responseBody}
}

type GetRequester struct {
	URL string
	M   Marshaller
}

func (r GetRequester) Request() Response {
	responseBody, err := GetHttpData(r.URL)
	if err != nil {
		return Response{Err: common.GenericError{Msg: r.URL, Extra: err}}
	}

	return Response{Body: responseBody}
}

func BulkRequest(requesters []Requester) BulkResponse {
	responsesChan := make(chan Response)
	var wg sync.WaitGroup

	for _, requester := range requesters {
		wg.Add(1)
		go func(r Requester) {
			defer wg.Done()
			responsesChan <- r.Request()
		}(requester)
	}

	go func() {
		wg.Wait()
		close(responsesChan)
	}()

	responses := make([]Response, 0)
	for ch := range responsesChan {
		responses = append(responses, ch)
	}

	return responses
}
