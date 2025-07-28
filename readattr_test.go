package fir

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/livefir/fir/internal/helper"
	"github.com/yosssi/gohtml"
	"golang.org/x/net/html"
)

func Test_query(t *testing.T) {
	type args struct {
		fi fileInfo
	}
	tests := []struct {
		name string
		args args
		want fileInfo
	}{

		{
			name: "no filters in event string and key is not present",
			args: args{
				fi: fileInfo{
					name: "test.html",
					content: []byte(`<!DOCTYPE html> 
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>`),
				},
			},
			want: fileInfo{
				name: "test.html",
				content: []byte(`<!DOCTYPE html> 
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>`),
				eventTemplates: eventTemplates{
					"event:ok": eventTemplate{
						"tmpl1": struct{}{},
						"-":     struct{}{},
					},
				}},
		},
		{
			name: "same event multiple elements: no filters in event string and key is not present",
			args: args{
				fi: fileInfo{
					name: "test.html",
					content: []byte(`<!DOCTYPE html> 
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>
					`),
				},
			},
			want: fileInfo{
				name: "test.html",
				content: []byte(`<!DOCTYPE html> 
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>
					`),
				eventTemplates: eventTemplates{
					"event:ok": eventTemplate{
						"tmpl1": struct{}{},
						"-":     struct{}{},
					},
				}},
		},
		{
			name: "multiple events, multiple templates, multiple elements: no filters in event string and key is not present",
			args: args{
				fi: fileInfo{
					name: "test.html",
					content: []byte(`<!DOCTYPE html> 
					<div
						 @fir:event1:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>
					<div
						 @fir:event2:ok::tmpl2=""
						 @fir:event:ok=""> 
					</div>
					`),
				},
			},
			want: fileInfo{
				name: "test.html",
				content: []byte(`<!DOCTYPE html> 
					<div
						 @fir:event1:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>
					<div
						 @fir:event2:ok::tmpl2=""
						 @fir:event:ok=""> 
					</div>
					`),
				eventTemplates: eventTemplates{
					"event:ok": eventTemplate{
						"-": struct{}{},
					},
					"event1:ok": eventTemplate{
						"tmpl1": struct{}{},
					},
					"event2:ok": eventTemplate{
						"tmpl2": struct{}{},
					},
				}},
		},
		{
			name: "filters in event string and key is present",
			args: args{
				fi: fileInfo{
					name: "test.html",
					content: []byte(`<!DOCTYPE html> 
					<div fir-key="1"
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok::tmpl2=""  
						 @fir:event:ok::tmpl2="" 
						 @fir:event:ok=""
						 @fir:[event1:ok,event2:ok]::tmpl3="console.log('hello')"> 
					</div>`),
				},
			},
			want: fileInfo{
				name: "test.html",
				content: []byte(`<!DOCTYPE html> 
					<div fir-key="1"
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok::tmpl2=""  
						 @fir:event:ok::tmpl2="" 
						 @fir:event:ok=""
						 @fir:[event1:ok,event2:ok]::tmpl3="console.log('hello')"> 
					</div>`),
				eventTemplates: eventTemplates{
					"event:ok": eventTemplate{
						"tmpl1": struct{}{},
						"tmpl2": struct{}{},
						"-":     struct{}{},
					},
					"event1:ok": eventTemplate{
						"tmpl3": struct{}{},
					},
					"event2:ok": eventTemplate{
						"tmpl3": struct{}{},
					},
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readAttributes(tt.args.fi)
			if !reflect.DeepEqual(got.eventTemplates, tt.want.eventTemplates) {
				t.Errorf("eventTemplates = %v, want %v", got.eventTemplates, tt.want.eventTemplates)
			}

			gotNode, err := html.Parse(bytes.NewReader(got.content))
			if err != nil {
				t.Fatalf("error parsing html: %v", err)
			}
			wantNode, err := html.Parse(bytes.NewReader(tt.want.content))
			if err != nil {
				t.Fatalf("error parsing html: %v", err)
			}

			if err := helper.AreNodesDeepEqual(gotNode, wantNode); err != nil {
				t.Fatalf("\nerr: %v \ngot \n %v \n want \n %v", err, gohtml.Format(string(helper.HtmlNodeToBytes(gotNode))), gohtml.Format(string(helper.HtmlNodeToBytes(wantNode))))
			}
		})
	}
}

func TestGetEventFilter(t *testing.T) {
	tests := []struct {
		input          string
		expectedBefore string
		expectedValues []string
		expectedAfter  string
		err            error
	}{
		{
			input:          "SomeText[value1:ok,value2:pending,value3:error]moreText",
			expectedBefore: "SomeText",
			expectedValues: []string{"value1:ok", "value2:pending", "value3:error"},
			expectedAfter:  "moreText",
		},
		{
			input:          "[value1:ok,value2:pending,value3:error]moreText",
			expectedBefore: "",
			expectedValues: []string{"value1:ok", "value2:pending", "value3:error"},
			expectedAfter:  "moreText",
		},
		{
			input:          "[value1:ok,value1-update:ok,value-update:pending,value3:error]",
			expectedBefore: "",
			expectedValues: []string{"value1:ok", "value1-update:ok", "value-update:pending", "value3:error"},
			expectedAfter:  "",
		},
		{
			input:          "SomeText:[value1:ok,value2:pending,value3:error]",
			expectedBefore: "SomeText:",
			expectedValues: []string{"value1:ok", "value2:pending", "value3:error"},
			expectedAfter:  "",
		},
		{
			input:          "SomeText:[value:ok]::moreText",
			expectedBefore: "SomeText:",
			expectedValues: []string{"value:ok"},
			expectedAfter:  "::moreText",
		},
		{
			input:          "SomeText[]moreText",
			expectedBefore: "SomeText",
			expectedValues: nil,
			expectedAfter:  "moreText",
			err:            ErrorEventFilterFormat,
		},
		{
			input:          "fir:event:ok::tmpl",
			expectedBefore: "",
			expectedValues: nil,
			expectedAfter:  "",
		},
		{
			input:          "SomeText[invalidFormat]moreText",
			expectedBefore: "",
			expectedValues: nil,
			expectedAfter:  "",
			err:            ErrorEventFilterFormat,
		},
	}

	for _, test := range tests {
		ef, err := getEventFilter(test.input)
		if err != test.err {
			t.Fatalf("Failed to parse event filter for input: %s, error: = %v", test.input, err)
		}

		if ef == nil && test.expectedValues == nil {
			continue
		}

		if ef.BeforeBracket != test.expectedBefore {
			t.Errorf("BeforeBracket mismatch for input: %s, expected: %s, got: %s", test.input, test.expectedBefore, ef.BeforeBracket)
		}

		if !reflect.DeepEqual(ef.Values, test.expectedValues) {
			t.Errorf("Values mismatch for input: %s, expected: %v, got: %v", test.input, test.expectedValues, ef.Values)
		}

		if ef.AfterBracket != test.expectedAfter {
			t.Errorf("AfterBracket mismatch for input: %s, expected: %s, got: %s", test.input, test.expectedAfter, ef.AfterBracket)
		}

	}
}
