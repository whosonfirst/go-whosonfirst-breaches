prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep
	if test -d src/github.com/whosonfirst/go-whosonfirst-breaches; then rm -rf src/github.com/whosonfirst/go-whosonfirst-breaches; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-breaches
	cp breaches.go src/github.com/whosonfirst/go-whosonfirst-breaches/

deps:   self
	go get -u "github.com/whosonfirst/go-whosonfirst-rtree"
	go get -u "github.com/whosonfirst/go-whosonfirst-utils"
	go get -u "github.com/whosonfirst/go-whosonfirst-log"
	go get -u "github.com/akavel/polyclip-go"

fmt:
	go fmt *.go
	go fmt cmd/*.go

bin:	self fmt
	go build -o bin/wof-breaches cmd/wof-breaches.go
	go build -o bin/wof-breach-server cmd/wof-breach-server.go
	go build -o bin/wof-clipping cmd/wof-clipping.go

ca:
	./bin/wof-clipping -data /usr/local/mapzen/whosonfirst-data/data/ -clipping 85633041 -subject 85633793 -loglevel info

mtl:
	./bin/wof-clipping -data /usr/local/mapzen/whosonfirst-data/data/ -subject 101736451 -clipping 101736545 -loglevel info
	./bin/wof-clipping -data /usr/local/mapzen/whosonfirst-data/data/ -clipping 101736451 -subject 101736545 -loglevel info

lengua:
	./bin/wof-clipping -data /usr/local/mapzen/whosonfirst-data/data/ -clipping 102112179 -subject 85887469	
	./bin/wof-clipping -data /usr/local/mapzen/whosonfirst-data/data/ -subject 102112179 -clipping 85887469	
