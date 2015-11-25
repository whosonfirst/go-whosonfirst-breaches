prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep
	if test -d src/github.com/whosonfirst/go-whosonfirst-breaches; then rm -rf src/github.com/whosonfirst/go-whosonfirst-breaches; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-breaches
	cp breaches.go src/github.com/whosonfirst/go-whosonfirst-breaches/

deps:   self
	go get -u "github.com/whosonfirst/go-whosonfirst-rtree"
	go get -u "github.com/akavel/polyclip-go"
