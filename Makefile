TAG="\n\n\033[0;32m\#\#\# "
END=" \#\#\# \033[0m\n"

test:
	@echo $(TAG)Make Test$(END)
	cd upyun && go test
	@echo
