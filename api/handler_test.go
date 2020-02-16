package api

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/globusdigital/feature-toggles/toggle"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	flags1 = []toggle.Flag{
		{Name: "flag1", ServiceName: "svc1", RawValue: "t", Value: true},
		{Name: "flag2", ServiceName: "", RawValue: "1", Value: true},
		{Name: "flag3", ServiceName: "svc1", RawValue: "f"},
		{Name: "flag4", ServiceName: "svc2", RawValue: "some string"},
	}

	strFlags1 = `[{"name": "flag10", "serviceName": "svc1", "rawValue": "1", "value": true}, {"name": "flag11", "serviceName": "svc2", "rawValue": "1", "value": true}]`
	strFlags2 = `[{"name": "flag10", "serviceName": "svc1", "rawValue": "1", "value": true}, {"name": "flag11", "serviceName": "", "rawValue": "raw string"}]`
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name string

		method string
		url    string
		body   string

		serviceName string
		flags       []toggle.Flag
		flagsErr    error

		saveInitial  bool
		flagsSaveErr error

		sendErr error

		wantCode int
		want     interface{}
	}{
		{name: "invalid url", url: "/", wantCode: 404},
		{name: "get all flags err", url: "/flags", flagsErr: errors.New("flags get"), wantCode: 500},
		{name: "get all flags", url: "/flags", flags: flags1, wantCode: 200, want: flags1},

		{name: "get flags svc1 err", url: "/flags/svc1", serviceName: "svc1", flagsErr: errors.New("flags get"), wantCode: 500},
		{name: "get flags svc1", url: "/flags/svc1", serviceName: "svc1", flags: flags1[:3], wantCode: 200, want: flags1[:3]},

		{name: "save flags svc1, no body", method: "POST", url: "/flags/svc1", serviceName: "svc1", wantCode: 400},
		{name: "save flags svc1 invalid", method: "POST", url: "/flags/svc1", body: strFlags1, serviceName: "svc1", wantCode: 400},
		{name: "save flags svc1 save err", method: "POST", url: "/flags/svc1", body: strFlags2, serviceName: "svc1", flagsSaveErr: errors.New("save err"), wantCode: 500},
		{name: "save flags svc1 send err", method: "POST", url: "/flags/svc1", body: strFlags2, serviceName: "svc1", sendErr: errors.New("save err"), wantCode: 500},
		{name: "save flags svc1", method: "POST", url: "/flags/svc1", body: strFlags2, serviceName: "svc1", wantCode: 204},

		{name: "delete flags svc1, no body", method: "DELETE", url: "/flags/svc1", serviceName: "svc1", wantCode: 400},
		{name: "delete flags svc1 invalid", method: "DELETE", url: "/flags/svc1", body: strFlags1, serviceName: "svc1", wantCode: 400},
		{name: "delete flags svc1 save err", method: "DELETE", url: "/flags/svc1", body: strFlags2, serviceName: "svc1", flagsSaveErr: errors.New("save err"), wantCode: 500},
		{name: "delete flags svc1 send err", method: "DELETE", url: "/flags/svc1", body: strFlags2, serviceName: "svc1", sendErr: errors.New("save err"), wantCode: 500},
		{name: "delete flags svc1", method: "DELETE", url: "/flags/svc1", body: strFlags2, serviceName: "svc1", wantCode: 204},

		{name: "save initial flags svc1, no body", method: "POST", url: "/flags/svc1/initial", serviceName: "svc1", saveInitial: true, wantCode: 400},
		{name: "save initial flags svc1 invalid", method: "POST", url: "/flags/svc1/initial", body: strFlags1, serviceName: "svc1", saveInitial: true, wantCode: 400},
		{name: "save initial flags svc1 save err", method: "POST", url: "/flags/svc1/initial", body: strFlags2, serviceName: "svc1", flagsSaveErr: errors.New("save err"), saveInitial: true, wantCode: 500},
		{name: "save initial flags svc1", method: "POST", url: "/flags/svc1/initial", body: strFlags2, serviceName: "svc1", saveInitial: true, flags: flags1, wantCode: 200, want: flags1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store, bus := NewMockStore(ctrl), NewMockBus(ctrl)
			store.EXPECT().Get(gomock.Any(), gomock.Eq(tt.serviceName)).AnyTimes().Return(tt.flags, tt.flagsErr)
			store.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Eq(tt.saveInitial)).AnyTimes().Return(tt.flagsSaveErr)

			bus.EXPECT().Send(gomock.Any(), gomock.Any()).AnyTimes().Return(tt.sendErr)

			w, r := httptest.NewRecorder(), httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))

			Handler(store, bus).ServeHTTP(w, r)

			a := assert.New(t)
			a.Equal(tt.wantCode, w.Code, w.Body.String())

			if w.Code != 200 {
				return
			}

			b, err := json.Marshal(tt.want)
			a.NoError(err)
			a.Equal(string(b), w.Body.String())
		})
	}
}
