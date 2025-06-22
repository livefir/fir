package fir

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/livefir/fir/internal/firattr"
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

			if err := areNodesDeepEqual(gotNode, wantNode); err != nil {
				t.Fatalf("\nerr: %v \ngot \n %v \n want \n %v", err, gohtml.Format(string(firattr.HTMLNodeToBytes(gotNode))), gohtml.Format(string(firattr.HTMLNodeToBytes(wantNode))))
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
			err:            firattr.ErrorEventFilterFormat,
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
			err:            firattr.ErrorEventFilterFormat,
		},
	}

	for _, test := range tests {
		ef, err := firattr.GetEventFilter(test.input)
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

func areNodesDeepEqual(node1, node2 *html.Node) error {
	if node1 == nil && node2 == nil {
		return fmt.Errorf("both nodes are nil")
	}

	if node1 == nil || node2 == nil {
		return fmt.Errorf("one of the nodes is nil")
	}

	if node1.Type != node2.Type {
		return fmt.Errorf("node types are not equal (%v != %v)", node1.Type, node2.Type)
	}

	if removeSpace(node1.Data) != removeSpace(node2.Data) {
		return fmt.Errorf("node data is not equal (%s != %s)", node1.Data, node2.Data)
	}

	if err := areAttributesEqual(node1.Attr, node2.Attr); err != nil {
		return err
	}

	c1 := node1.FirstChild
	c2 := node2.FirstChild

	for c1 != nil && c2 != nil {
		if err := areNodesDeepEqual(c1, c2); err != nil {
			return err
		}

		c1 = c1.NextSibling
		c2 = c2.NextSibling

	}

	if c1 != nil && c1.DataAtom.String() != "" {
		return fmt.Errorf("node1 has extra child: atom: %v, val: %v\n", c1.DataAtom, string(firattr.HTMLNodeToBytes(c1)))
	}
	if c2 != nil && c2.DataAtom.String() != "" {
		return fmt.Errorf("node2 has extra child: atom: %v, val: %v\n", c2.DataAtom, string(firattr.HTMLNodeToBytes(c2)))
	}

	return nil
}

func areAttributesEqual(attr1, attr2 []html.Attribute) error {

	attr1Map := make(map[string]string)
	for _, a := range attr1 {
		attr1Map[a.Key] = a.Val
	}

	attr2Map := make(map[string]string)
	for _, a := range attr2 {
		attr2Map[a.Key] = a.Val
	}

	for k, v := range attr1Map {
		if k == "class" {
			if err := areClassesEqual(v, attr2Map["class"]); err != nil {
				return err
			}
		} else {
			val, ok := attr2Map[k]
			if !ok {
				return fmt.Errorf("attr %v is not present in attr2Map %+v", k, attr2Map)
			}
			if val != v {
				return fmt.Errorf("attr %v has different values: %v != %v", k, val, v)
			}
		}

		delete(attr2Map, k)
	}

	if len(attr2Map) > 0 {
		return fmt.Errorf("attr2Map has extra attributes: %v", attr2Map)
	}

	return nil
}

func areClassesEqual(class1, class2 string) error {
	classSet1 := strings.Fields(class1)
	classSet2 := strings.Fields(class2)

	classMap := make(map[string]bool)
	for _, class := range classSet1 {
		classMap[class] = true
	}

	for _, class := range classSet2 {
		_, ok := classMap[class]
		if !ok {
			return fmt.Errorf("class %v is not present in classSet1", class)
		}
	}

	return nil
}
