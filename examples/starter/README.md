# starter

```bash
brew install gh
gh repo create myapp --template goliveview/starter
cd myapp
make install
# replace goliveview-starter with your app name
go get github.com/piranha/goreplace
$(go env GOPATH)/bin/goreplace goliveview-starter -r myapp
git add . && git commit -m "replace goliveview-starter"
cp env.dev env.local
make watch
```