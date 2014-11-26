TAG="\n\n\033[0;32m\#\#\# "
END=" \#\#\# \033[0m\n"

test: prepare start 

prepare:
	@echo $(TAG)Environment Variable$(END)
	export UPYUN_BUCKET=sdkfile
	export UPYUN_USERNAME=tester
	export UPYUN_PASSWORD=grjxv2mxELR3
	@echo

start:
	@echo $(TAG)Make Test$(END)
	cd tests && go test
	@echo
