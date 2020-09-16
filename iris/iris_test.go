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

type testType struct {
	Name    string `json:"name"`
	Company string `json:"company"`
}

// func getJSendError(w http.ResponseWriter, req *http.Request) {

// }

// func getJSendSuccess(w http.ResponseWriter, req *http.Request) {
// 	j, err := json.Marshal(map[string]interface{}{"data": map[string]interface{}{"name": "Gbolahan"}})
// 	if err != nil {
// 		panic(err)
// 	}

// 	_, err = w.Write(j)
// 	if err != nil {
// 		panic(err)
// 	}
// }

func TestGetResponse(t *testing.T) {
	t.Run("should get decoded response", func(t *testing.T) {
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

		var tType testType
		_, err = GetResponse(&res, &tType)
		if err != nil {
			t.Fatal(errors.Wrap(err, "error getting response"))
		}

		fmt.Printf("")
	})
}
