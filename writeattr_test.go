package fir

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yosssi/gohtml"
	"golang.org/x/net/html"
)

func Test_addAttributes(t *testing.T) {

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "key is present but no @ or x-on is present",
			input: `
					<div key="parent-key">
							<p>Hello, World!</p>
							<div>
								<span>Inner element</span>
							</div>
					</div>
				`,
			want: `
					<div key="parent-key">
							<p>Hello, World!</p>
							<div>
								<span>Inner element</span>
							</div>
						</div>
				`,
		},
		{
			name: "key is present and @ is present but child already has key",
			input: `
					<div key="parent-key">
						<p>Hello, World!</p>
						<div>
							<span key="" @click="console.log()">Inner element</span>
						</div>
					</div>
				`,
			want: `
					<div key="parent-key">
						<p>Hello, World!</p>
						<div>
							<span key="" @click="console.log()">Inner element</span>
						</div>
					</div>
				`,
		},
		{
			name: "key is present and @ is present and child does not have key",
			input: `
					<div key="parent-key">
						<p>Hello, World!</p>
						<div x-on:click="doSomething()">
							<span>Inner element</span>
						</div>
					</div>
		 `,
			want: `
					<div key="parent-key">
						<p>Hello, World!</p>
						<div x-on:click="doSomething()" key="parent-key">
							<span>Inner element</span>
						</div>
					</div>
				`,
		},
		{
			name: "two elements: key is present and @ is present and child does not have key",
			input: `
					<div key="parent-key">
						<p>Hello, World!</p>
						<div x-on:click="doSomething()">
							<span>Inner element</span>
						</div>
					</div>
					<div key="parent-key-2">
						<p>Hello, World!</p>
						<div x-on:click="doSomething()">
							<span>Inner element</span>
						</div>
					</div>
				`,
			want: `
					<div key="parent-key">
						<p>Hello, World!</p>
						<div x-on:click="doSomething()" key="parent-key">
							<span >Inner element</span>
						</div>
					</div>
					<div key="parent-key-2">
						<p>Hello, World!</p>
						<div x-on:click="doSomething()" key="parent-key-2">
							<span >Inner element</span>
						</div>
					</div>
				`,
		},
		{
			name: "add key attribute to nested children",
			input: `
					<div key="1"> 
						<div @fir:event:ok="" > </div>
						<div> <button @click="">  </button></div>
						<div> <div> <div> <input @change="" > </div> </div> </div>
						<form @submit=""> </form>
					</div>
					`,

			want: `
					<div key="1"> 
						<div key="1" @fir:event:ok="" class="fir-event-ok--1"> </div>
						<div> <button @click="" key="1">  </button> </div>
						<div> <div> <div> <input @change="" key="1"> </div> </div> </div>
						<form @submit="" key="1"> </form>
					</div>
					`,
		},
		{
			name: "no filters in event string and key is not present",
			input: `
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok="">
					</div>
					`,

			want: `
					<div class="fir-event-ok--tmpl1 fir-event-ok"
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok="">
					</div>
						 `,
		},
		{
			name: "same event multiple elements: no filters in event string and key is not present",
			input: `
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok="">
					</div>
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok="">
					</div>
					`,

			want: `
					<div class="fir-event-ok--tmpl1 fir-event-ok" 
						@fir:event:ok::tmpl1="" 
						@fir:event:ok="">
					</div>
					<div class="fir-event-ok--tmpl1 fir-event-ok"
						@fir:event:ok::tmpl1=""
						@fir:event:ok="">
					</div>
				`,
		},

		{
			name: "multiple events, multiple templates, multiple elements: no filters in event string and key is not present",
			input: `
					<div
						 @fir:event1:ok::tmpl1=""
						 @fir:event:ok="">
					</div>
					<div
						 @fir:event2:ok::tmpl2=""
						 @fir:event:ok="">
					</div>
					`,

			want: `
						<div class="fir-event1-ok--tmpl1 fir-event-ok"
							@fir:event1:ok::tmpl1=""
							@fir:event:ok="">
						 </div>
						 <div class="fir-event2-ok--tmpl2 fir-event-ok"
							@fir:event2:ok::tmpl2=""
							@fir:event:ok="">
						 </div>
						 `,
		},
		{
			name: "filters in event string and key is present",
			input: `
					<div key="1"
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok::tmpl2=""
						 @fir:event:ok::tmpl2=""
						 @fir:event:ok=""
						 @fir:[event1:ok,event2:ok]::tmpl3="console.log('hello')"
						 @fir:event2:ok="">
					</div>
					`,

			want: `
					<div key="1"
						 class="fir-event-ok--tmpl1--1 fir-event-ok--tmpl2--1 fir-event-ok--1 fir-event1-ok--tmpl3--1 fir-event2-ok--tmpl3--1 fir-event2-ok--1"
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok::tmpl2=""
						 @fir:event:ok::tmpl2=""
						 @fir:event:ok=""
						 @fir:event1:ok::tmpl3="console.log('hello')"
						 @fir:event2:ok::tmpl3="console.log('hello')"
						 @fir:event2:ok="">
					</div>
						 `,
		},
		{
			name: "filters in event string and no key is present",
			input: `
					<div
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok::tmpl2=""
						 @fir:event:ok::tmpl2=""
						 @fir:[event1:ok,event2:ok]::tmpl3="console.log('hello')"
						 @fir:event2:ok="">
					</div>
					`,

			want: `
					<div
						 class="fir-event-ok--tmpl1 fir-event-ok--tmpl2 fir-event1-ok--tmpl3 fir-event2-ok--tmpl3 fir-event2-ok"
						 @fir:event:ok::tmpl1=""
						 @fir:event:ok::tmpl2=""
						 @fir:event:ok::tmpl2=""
						 @fir:event1:ok::tmpl3="console.log('hello')"
						 @fir:event2:ok::tmpl3="console.log('hello')"
						 @fir:event2:ok="">
					</div>
						 `,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			want, err := html.Parse(strings.NewReader(test.want))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}
			input := addAttributes([]byte(test.input))
			got, err := html.Parse(bytes.NewReader(input))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			if err := areNodesDeepEqual(got, want); err != nil {
				t.Fatalf("\nerr: %v \ngot \n %v \n want \n %v", err, gohtml.Format(string(htmlNodeToBytes(got))), gohtml.Format(string(htmlNodeToBytes(want))))
			}
		})
	}
}
