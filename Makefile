.PHONY : clean release

clean:
	go mod vendor
	go mod tidy

release: clean
	standard-version
	git push --follow-tags origin master