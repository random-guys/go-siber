package iris

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"
	"syreclabs.com/go/faker"
)

type jSMock struct {
	Data interface{}
}

func TestGetResponse(t *testing.T) {
	t.Run("It should decode response", func(t *testing.T) {
		name := faker.Name().FirstName()
		company := faker.Company().Name()

		b, err := json.Marshal(map[string]interface{}{
			"data": map[string]interface{}{"name": name, "company": company},
		})
		if err != nil {
			panic(err)
		}

		res := http.Response{
			Body: ioutil.NopCloser(bytes.NewBuffer(b)),
		}

		var data jSMock
		err = GetResponse(&res, &data)
		if err != nil {
			t.Fatal(errors.Wrap(err, "error getting response"))
		}

		fmt.Printf("here's the response: %#v\n", data)
	})

	t.Run("It should decode string response", func(t *testing.T) {
		b, err := json.Marshal(map[string]interface{}{
			"data": "response",
		})
		if err != nil {
			panic(err)
		}

		res := http.Response{
			Body: ioutil.NopCloser(bytes.NewBuffer(b)),
		}

		var data jSMock
		err = GetResponse(&res, &data)
		if err != nil {
			t.Fatal(errors.Wrap(err, "error getting response"))
		}

		fmt.Printf("here's the response: %#v\n", data)
	})
}
