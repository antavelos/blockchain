package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/antavelos/blockchain/pkg/common"
)

func GetHttpData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	return body, err
}

func PostHttpData(url string, data []byte) ([]byte, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}

	return body, nil
}

type Marshaller interface {
	Unmarshal([]byte) (any, error)
}

type ObjectMarshaller struct {
	Many bool
}

type Response struct {
	Body any
	Err  error
}

type BulkResponse []Response

func (br BulkResponse) errorResponses() BulkResponse {
	return common.Filter(br, func(r Response) bool {
		return r.Err != nil
	})
}

func (br BulkResponse) ErrorsRatio() float64 {
	return float64(len(br.errorResponses())) / float64(len(br))
}

func (br BulkResponse) ErrorStrings() []string {
	return common.Map(br, func(r Response) string {
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
		return Response{Err: fmt.Errorf("%v: %v", r.URL, err.Error())}
	}

	if r.M == nil {
		return Response{Body: responseBody}
	}

	unmarshalledBody, err := r.M.Unmarshal(responseBody)
	if err != nil {
		return Response{Err: fmt.Errorf("%v: failed to unmarshal response: %v", r.URL, err.Error())}
	}

	return Response{Body: unmarshalledBody}
}

type GetRequester struct {
	URL string
	M   Marshaller
}

func (r GetRequester) Request() Response {
	responseBody, err := GetHttpData(r.URL)
	if err != nil {
		return Response{Err: fmt.Errorf("%v: %v", r.URL, err.Error())}
	}

	if r.M == nil {
		return Response{Body: responseBody}
	}

	unmarshalledBody, err := r.M.Unmarshal(responseBody)
	if err != nil {
		return Response{Err: fmt.Errorf("%v: failed to unmarshal response: %v", r.URL, err.Error())}
	}

	return Response{Body: unmarshalledBody}
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
