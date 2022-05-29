# starter

```bash
brew install gh
gh repo create myapp --template pineview/starter
cd myapp
make install
# replace pineview-starter with your app name
go get github.com/piranha/goreplace
$(go env GOPATH)/bin/goreplace pineview-starter -r myapp
git add . && git commit -m "replace pineview-starter"
cp env.dev env.local
make watch
```
