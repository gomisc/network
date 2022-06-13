package nethttp

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_request_Params(t *testing.T) {
	type fields struct {
		r   *http.Request
		cli *http.Client
		err error
	}

	type args struct {
		params interface{}
	}

	type formParams struct {
		StringParam string   `form:"param1"`
		ArrayParam  []string `form:"param2"`
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   R
	}{
		{
			name: "params with empty params",
			fields: fields{
				r: &http.Request{URL: &url.URL{}},
			},
			args: args{
				params: map[string]interface{}{},
			},
			want: func() R {
				return &request{
					r: &http.Request{URL: &url.URL{}},
				}
			}(),
		},
		{
			name: "params with single values",
			fields: fields{
				r: &http.Request{URL: &url.URL{}},
			},
			args: args{
				params: map[string]interface{}{
					"param1": "value1",
					"param2": "value2",
				},
			},
			want: func() R {
				return &request{
					r: &http.Request{URL: &url.URL{RawQuery: "param1=value1&param2=value2"}},
				}
			}(),
		},
		{
			name: "params with array values",
			fields: fields{
				r: &http.Request{URL: &url.URL{}},
			},
			args: args{
				params: map[string][]interface{}{
					"param1": {1, 2, 3},
					"param2": {"one", "two", "three"},
				},
			},
			want: func() R {
				return &request{
					r: &http.Request{URL: &url.URL{RawQuery: "param1=1&param1=2&param1=3&param2=one&param2=two&param2=three"}},
				}
			}(),
		},
		{
			name: "form data params",
			fields: fields{
				r: &http.Request{URL: &url.URL{}},
			},
			args: args{
				params: &formParams{
					StringParam: "value1",
					ArrayParam:  []string{"one", "two", "three"},
				},
			},
			want: func() R {
				return &request{
					r: &http.Request{URL: &url.URL{RawQuery: "param1=value1&param2=one&param2=two&param2=three"}},
				}
			}(),
		},
	}

	for i := 0; i < len(tests); i++ {
		tt := tests[i]

		t.Run(tt.name, func(t *testing.T) {
			r := &request{
				r:   tt.fields.r,
				cli: tt.fields.cli,
				err: tt.fields.err,
			}

			got := r.Params(tt.args.params)
			ow, og := tt.want.(*request), got.(*request)

			wuu, err := url.QueryUnescape(ow.r.URL.RequestURI())
			assert.NoError(t, err)
			guu, err := url.QueryUnescape(og.r.URL.RequestURI())
			assert.NoError(t, err)

			if !assert.EqualValues(t, wuu, guu) {
				t.Fail()
			}
		})
	}
}
