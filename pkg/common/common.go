package common

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

var InfoLogger *log.Logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
var ErrorLogger *log.Logger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

type GenericError struct {
	Msg   string
	Extra error
}

func (e GenericError) Error() string {
	msg := e.Msg
	if e.Extra != nil {
		msg = fmt.Sprintf("%v: %v", e.Msg, e.Extra.Error())
	}
	ErrorLogger.Println(msg)
	return msg
}

func Map[T, R any](data []T, f func(T) R) []R {

	res := make([]R, 0, len(data))

	for _, e := range data {
		res = append(res, f(e))
	}

	return res
}

func Filter[T any](data []T, f func(T) bool) []T {

	res := make([]T, 0, len(data))

	for _, e := range data {
		b := f(e)
		if b {
			res = append(res, e)
		}
	}

	return res
}

func ErrorsToStrings(err []error) []string {
	return Map(
		Filter(err, func(e error) bool {
			return e != nil
		}),
		func(err error) string {
			return err.Error()
		})

}

func GetRandomInt(numRange int) int {
	if numRange == 0 {
		return 0
	}

	rand.Seed(time.Now().UnixNano())

	return rand.Intn(numRange)
}

func GetRandomFloat(min, max float64) float64 {
	rand.Seed(time.Now().UnixNano())

	return min + rand.Float64()*(max-min)
}
