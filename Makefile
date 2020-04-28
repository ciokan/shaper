.PHONY : clean release

clean:
	go mod vendor
	go mod tidy

release: clean
	standard-version
	git push --follow-tags origin master

#	only run it to debug things locally
#	use --snapshot to build local
#	(will produce files that cannot be uploaded to bintray)
gorelease:
	docker run --rm --privileged \
		-v ${PWD}:/go/src/github.com/ciokan/shaper \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /go/src/github.com/dnsadblock/proxy \
		--env-file .env \
		goreleaser/goreleaser release --rm-dist --snapshot --skip-publish