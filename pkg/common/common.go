package common

import (
	"log"
	"os"
)

var InfoLogger *log.Logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
var ErrorLogger *log.Logger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

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
