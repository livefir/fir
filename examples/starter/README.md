# starter

```bash
brew install gh
gh repo create myapp --template fir/starter
cd myapp
make install
# replace fir-starter with your app name
go get github.com/piranha/goreplace
$(go env GOPATH)/bin/goreplace fir-starter -r myapp
git add . && git commit -m "replace fir-starter"
cp env.dev env.local
make watch
```
