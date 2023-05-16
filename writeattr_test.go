package fir

import (
	"testing"
)

func Test_transform(t *testing.T) {

	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		// TODO: Add test cases.
		{
			name: "add key attribute to children",
			input: []byte(`<!DOCTYPE html> 
					<div key="1"> 
						<div> <button @click="">  </button> </div>
						<div> <div> <div> <input @change="" > </div> </div> </div>
						<form @submit=""> </form>
					</div>
					`),

			want: []byte(`<!DOCTYPE html> 
					<div key="1"> 
						<div @fir:event:ok="" key="1" > </div>
						<div> <button @click="" key="1">  </button> </div>
						<div> <div> <div> <input @change="" key="1"> </div> </div> </div>
						<form @submit="" key="1"> </form>
					</div>
					`),
		},
		{
			name: "no filters in event string and key is not present",
			input: []byte(`<!DOCTYPE html> 
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>`),

			want: []byte(`<!DOCTYPE html> 
					<div class="fir-event-ok--tmpl1 fir-event-ok" 
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok="" 
						 </div>`),
		},
		{
			name: "same event multiple elements: no filters in event string and key is not present",
			input: []byte(`<!DOCTYPE html> 
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>
					`),

			want: []byte(`<!DOCTYPE html> 
						<div class="fir-event-ok--tmpl1 fir-event-ok" 
							@fir:event:ok::tmpl1=""
							@fir:event:ok="" 
						 </div>
						 <div class="fir-event-ok--tmpl1 fir-event-ok" 
							@fir:event:ok::tmpl1=""
							@fir:event:ok="" 
						 </div>
						 `),
		},
		{
			name: "multiple events, multiple templates, multiple elements: no filters in event string and key is not present",
			input: []byte(`<!DOCTYPE html> 
					<div
						 @fir:event1:ok::tmpl1=""
						 @fir:event:ok=""> 
					</div>
					<div
						 @fir:event2:ok::tmpl2=""
						 @fir:event:ok=""> 
					</div>
					`),

			want: []byte(`<!DOCTYPE html> 
						<div class="fir-event1-ok--tmpl1 fir-event-ok" 
							@fir:event1:ok::tmpl1=""
							@fir:event:ok="" 
						 </div>
						 <div class="fir-event2-ok--tmpl2 fir-event-ok" 
							@fir:event2:ok::tmpl2=""
							@fir:event:ok="" 
						 </div>
						 `),
		},
		{
			name: "filters in event string and key is present",
			input: []byte(`<!DOCTYPE html> 
					<div key="1"
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok::tmpl2=""  
						 @fir:event:ok::tmpl2="" 
						 @fir:event:ok=""
						 @fir:[event1:ok,event2:ok]::tmpl3="console.log('hello')"> 
					</div>`),

			want: []byte(`<!DOCTYPE html> 
					<div key="1" 
						 class="fir-event-ok--tmpl1--1 fir-event-ok--tmpl2--1 fir-event-ok--1 fir-event1-ok--tmpl3--1 fir-event2-ok--tmpl3--1" 
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok::tmpl2=""  
						 @fir:event:ok::tmpl2="" 
						 @fir:event:ok=""
						 @fir:event1:ok::tmpl3="console.log('hello')"
						 @fir:event2:ok::tmpl3="console.log('hello')"> 
						 </div>`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addAttributes(tt.input)
			if !areHTMLStringsEqual(t, got, tt.want) {
				t.Errorf("html \n %v, \n want \n %v", string(got), string(tt.want))
			}
		})
	}
}
