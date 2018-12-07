SHELL := /bin/bash
################################################################
# Makefile for build, test, and publish project related files
#
# Run examples (see more commands in "help" target):
# $ make	 	 -- creates the debian package.
# $ make build   -- builds binary without creating the debian package.
# $ make clean   -- removes all project related built files.
# $ make test 	 -- runs go test for all packages.
################################################################

#########################################
# Package Metadata
#########################################
REPO_NAME=fm-app-go-template
PROJECT_NAME=$(REPO_NAME)
VERSION=$(shell git describe --tags --always)
ARCH=amd64
CATEGORY=fusemail
VENDOR=fusemail
MAINTAINER=developers@fusemail.com
DESCRIPTION='Service description goes here.'
HOMEPAGE=https://bitbucket.org/fusemail/$(PROJECT_NAME)
TARGET=$(PROJECT_NAME)_$(VERSION)_$(ARCH).deb
GITHASH=$(shell git rev-parse --short HEAD)
GITBRANCH=$(shell git rev-parse --abbrev-ref HEAD | grep -oE "^([A-Z]+-[0-9]+|master)" | tr '[:upper:]' '[:lower:]')
BUILDDATE=$(shell date '+%Y-%m-%d_%H:%M:%S_%Z')
SYSPKG=bitbucket.org/$(VENDOR)/$(PROJECT_NAME)/vendor/bitbucket.org/fusemail/fm-lib-commons-golang/sys
INSTALLPATH=/usr/local/$(VENDOR)/$(PROJECT_NAME)
PREINSTALL=./conf/preinstall.sh
POSTINSTALL=./conf/post_install.sh
SUPERVISORDIR=/etc/supervisor/conf.d/
SETTINGSDIR=/etc/$(VENDOR)/$(PROJECT_NAME)
# Specify the datacenters where your job is suppose to run
DATACENTERS=fusepoint ireland london sweden1 sweden2 toronto denmark1

#########################################
# CI/CD Config 
#########################################
# registry UI available here: https://bamboo2203a.electric.net:10081/repository/
REGISTRY_ADDR=docker-registry.electric.net:10080

ifeq ($(GITBRANCH),master)
	BUILDTAG=$(VERSION)
	NOMAD_NAME=$(REPO_NAME)
else
	BUILDTAG=$(VERSION)-$(GITBRANCH)
	NOMAD_NAME=$(REPO_NAME)-$(GITBRANCH)-$(GITHASH)
endif

BUILD_ENV_VERSION=1.7-go1.10.3-stretch	# use latest REGISTRY_ADDR/fusemail/fm-utility-go-build tag
DEBIAN_VERSION=9.4		# use latest REGISTRY_ADDR/debian/debian tag

BAMBOO_SERVER=bamboo2203a.electric.net:8085

# checkout existing projects in [Bamboo](http://bamboo2203a.electric.net:8085/allPlans.action)
BAMBOO_PROJECT=Templates

# use the same project key in Babmoo (i.e. the resource name in project URL)
BAMBOO_PROJECT_KEY=TMPLTS

BAMBOO_REPO_NAME=$(REPO_NAME)
BAMBOO_PLAN_NAME=$(REPO_NAME)

# use the first letter of each word in REPO_NAME to make your own plan key
BAMBOO_PLAN_KEY=FAGT

#########################################
# Debian Packaging Options
#########################################
FPM=/usr/local/bin/fpm

.PHONY: build clean deploy jobspec help test test-v docker docker-build docker-artifacts docker-push docker-test docker-clean docker-run docker-run-stop bamboo-test bamboo-publish bamboo-clean rename rename-revert rename-clean

$(TARGET): build
	@mkdir -p build
	$(FPM) -n $(PROJECT_NAME)\
		-p build/$(PROJECT_NAME)_$(VERSION)_$(ARCH).deb \
		--version=$(VERSION)\
		--force\
		-s dir\
		-t deb\
		--deb-no-default-config-files\
		--vendor=$(VENDOR)\
		--category=$(CATEGORY)\
		--description=$(DESCRIPTION)\
		--url=$(HOMEPAGE)\
		--maintainer=$(MAINTAINER)\
		--before-install=$(PREINSTALL)\
		--after-install=$(POSTINSTALL)\
		--prefix=\
		build/$(PROJECT_NAME)=$(INSTALLPATH)/$(PROJECT_NAME)\
		conf/$(PROJECT_NAME)-prod.env=$(SETTINGSDIR)/$(PROJECT_NAME)-prod.env\
		conf/$(PROJECT_NAME).conf=$(SUPERVISORDIR)/$(PROJECT_NAME).conf

docs/dist/api.html: docs/raml/* docs/raml/examples/*
	@mkdir -p docs/dist
	raml2html docs/raml/api.raml > docs/dist/api.html

bindata.go: docs/dist/api.html README.md
	go-bindata README.md docs/...

build: docs/dist/api.html bindata.go *.go
	@mkdir -p build
	go build -o build/$(PROJECT_NAME) -ldflags "\
		-X $(SYSPKG).Version=$(VERSION) \
		-X $(SYSPKG).GitHash=$(GITHASH) \
		-X $(SYSPKG).BuildStamp=$(BUILDDATE) \
	"
	cp build/$(PROJECT_NAME) build/$(PROJECT_NAME)-$(VERSION)

test:
	go test `go list ./... | grep -v vendor`

test-v:
	go test -v `go list ./... | grep -v vendor`

#########################################
# Docker Targets
#########################################
docker: build/docker

build/docker:
	@mkdir -p build && \
	docker build --build-arg registry=$(REGISTRY_ADDR) --build-arg build_env_version=$(BUILD_ENV_VERSION) --build-arg debian_version=$(DEBIAN_VERSION) --build-arg project_name=$(PROJECT_NAME) -t fusemail/$(PROJECT_NAME):$(BUILDTAG) -t fusemail/$(PROJECT_NAME):latest . && \
	touch build/docker

docker-build: build/docker_build

build/docker_build:
	@mkdir -p build && \
	docker build --build-arg registry=$(REGISTRY_ADDR) --build-arg build_env_version=$(BUILD_ENV_VERSION) --build-arg debian_version=$(DEBIAN_VERSION) --build-arg project_name=$(PROJECT_NAME) --target builder -t fusemail/$(PROJECT_NAME):build . && \
	touch build/docker_build

docker-artifacts: docker-build
	-docker rm -f $(PROJECT_NAME)-extract
	docker create --name $(PROJECT_NAME)-extract fusemail/$(PROJECT_NAME):build
	docker cp $(PROJECT_NAME)-extract:/go/src/bitbucket.org/fusemail/$(PROJECT_NAME)/build .
	docker cp $(PROJECT_NAME)-extract:/go/src/bitbucket.org/fusemail/$(PROJECT_NAME)/docs .
	-docker rm -f $(PROJECT_NAME)-extract

docker-push: docker
	docker tag fusemail/$(PROJECT_NAME):$(BUILDTAG) $(REGISTRY_ADDR)/fusemail/$(PROJECT_NAME):$(BUILDTAG) && \
	docker push $(REGISTRY_ADDR)/fusemail/$(PROJECT_NAME):$(BUILDTAG) && \
	if [ "$(GITBRANCH)" = "master" ]; then \
		docker tag fusemail/$(PROJECT_NAME):latest $(REGISTRY_ADDR)/fusemail/$(PROJECT_NAME):latest && \
		docker push $(REGISTRY_ADDR)/fusemail/$(PROJECT_NAME):latest; \
	fi

docker-test: docker-build
	@mkdir -p test-result
	docker run --rm -v $(PWD)/test-result:/tmp/ fusemail/$(PROJECT_NAME):build go-test-to-junit.sh

docker-run: docker docker-run-stop
	docker run -d --rm --name test_run -p 8080:8080 fusemail/$(PROJECT_NAME):latest

docker-run-logs: docker-run
	docker logs -f test_run

docker-run-help: docker docker-run-stop
	docker run --rm --entrypoint ./$(PROJECT_NAME) fusemail/$(PROJECT_NAME):latest -h

docker-run-stop:
	-docker rm -f test_run

docker-test-clean:
	-docker rmi fusemail/$(PROJECT_NAME):build
	if [ -e "build/docker_build" ]; then rm build/docker_build; fi
	if [ -e "test-result/junit-test-report.xml" ]; then rm -rf test-result; fi

docker-push-clean:
	-docker rmi $(REGISTRY_ADDR)/fusemail/$(PROJECT_NAME):$(BUILDTAG)
	-docker rmi $(REGISTRY_ADDR)/fusemail/$(PROJECT_NAME):latest
	-docker rmi fusemail/$(PROJECT_NAME):$(BUILDTAG)
	-docker rmi fusemail/$(PROJECT_NAME):latest
	if [ -e "build/docker" ]; then rm build/docker; fi

docker-clean: docker-test-clean docker-push-clean

#########################################
# Bamboo Targets
#########################################
bamboo-build:
	docker build --no-cache -t bamboo/$(PROJECT_NAME):build bamboo-specs/

bamboo-test: bamboo-build
	docker run --rm bamboo/$(PROJECT_NAME):build test

bamboo-publish: bamboo-build
	docker run --rm bamboo/$(PROJECT_NAME):build -Ppublish-specs

bamboo-clean:
	if [ -d "bamboo-specs/bin" ]; then rm -rf bamboo-specs/bin; fi
	if [ -d "bamboo-specs/target" ]; then rm -rf bamboo-specs/target; fi

#########################################
# Deploy Targets by pushing to artifact server.
# Only applicable running directly from Bamboo.
#########################################
jobspec-tmpl:
	cd conf && \
	for dc in $(DATACENTERS); do \
		sed -e "s/devel/$$dc/g; s/\"-e\", \"dev\"/\"-e\", \"prod\"/" < jobspec-devel.hcl > "jobspec-$$dc.hcl"; \
	done

jobspec:
	@mkdir -p build
	for f in conf/*.hcl; do \
		name=$$(basename -- "$$f") && \
		sed -e "s/###BUILDTAG###/$(BUILDTAG)/; s/###NOMAD_NAME###/$(NOMAD_NAME)/; s/###REPO_NAME###/$(REPO_NAME)/" < "$$f" > "build/$$name"; \
	done

deploy-clean: docker-push-clean
	if [ -d "build/" ]; then rm -rf build; fi

deploy: docker-push jobspec
	cd build && artifact new_version $(PROJECT_NAME) $(BUILDTAG) || true && \
	for f in *.hcl; do \
		artifact push $(PROJECT_NAME) $(BUILDTAG) "$$f"; \
	done && \
	if [ "$(GITBRANCH)" = "master" ]; then \
		artifact tag $(PROJECT_NAME) $(BUILDTAG) latest; \
	fi && \
	for f in *.hcl; do \
		dc=$$(echo "$$f" | sed 's/^jobspec-\(.*\).hcl/\1/') && \
		if [ "$$dc" = "devel" ] || [ "$(GITBRANCH)" = "master" ]; then \
			curl -XPOST -d"{\"name\":\"$(PROJECT_NAME)\",\"version\":\"$(BUILDTAG)\"}" "http://localhost:4151/pub?topic=bamboo_$${dc}_deployment"; \
		fi; \
	done

#########################################
# Cleanup Target
#########################################
clean: docker-test-clean bamboo-clean deploy-clean
	rm -f docs/dist/*
	rm -f bindata.go

#########################################
# Rename Targets
#########################################
rename-java-specs:
	find bamboo-specs/ -name PlanSpec.java | xargs sed -i .bak 's/BAMBOO_SERVER = .*;/BAMBOO_SERVER = "http:\/\/$(BAMBOO_SERVER)";/; s/PROJECT_NAME = ".*"/PROJECT_NAME = "$(BAMBOO_PROJECT)"/; s/PROJECT_KEY = ".*"/PROJECT_KEY = "$(BAMBOO_PROJECT_KEY)"/; s/PLAN_KEY = ".*"/PLAN_KEY = "$(BAMBOO_PLAN_KEY)"/; s/REPO_NAME = ".*"/REPO_NAME = "$(REPO_NAME)"/'

rename-yaml-specs:
	find bamboo-specs/ -name bamboo.yml | xargs sed -i .bak 's/  key: .*/  key: $(BAMBOO_PROJECT_KEY)/; s/PROJECT_KEY = ".*"/PROJECT_KEY = "$(BAMBOO_PROJECT_KEY)"/; s/    key: .*/    key: $(BAMBOO_PLAN_KEY)/; s/    name: .*/    name: $(REPO_NAME)/'

rename-bamboo-specs: rename-java-specs rename-yaml-specs

rename-conf-files:
	find conf/ -name '*.*' | xargs sed -i .bak 's/fm-app-go-template/$(PROJECT_NAME)/g'
	for f in conf/fm-app-go-template*; do mv "$$f" $$(echo "$$f" | sed 's/fm-app-go-template/$(PROJECT_NAME)/g'); done

rename-docs:
	find docs/raml/ -name '*.raml' | xargs sed -i .bak 's/fm-app-go-template/$(PROJECT_NAME)/g'

rename-readme:
	find . -name 'README.template.md' | xargs sed -i .bak 's/fm-app-go-template/$(PROJECT_NAME)/g; s/$(PROJECT_NAME)\\)\./fm-app-go-template\\)\./'
	cp README.md TEMPLATE.md
	mv README.template.md README.md

rename: rename-bamboo-specs rename-conf-files rename-docs rename-readme

rename-revert-bamboo-specs:
	if [ -e "bamboo-specs/src/main/java/fusemail/PlanSpec.java.bak" ]; then \
		cd bamboo-specs/src/main/java/fusemail/; \
		PlanSpec.java.bak PlanSpec.java; \
	fi
	if [ -e "bamboo-specs/bamboo.yml.bak" ]; then \
		cd bamboo-specs/; \
		mv bamboo.yml.bak bamboo.yml; \
	fi

rename-revert-conf-files:
	for f in conf/$(PROJECT_NAME)*; do mv "$$f" $$(echo "$$f" | sed 's/$(PROJECT_NAME)/fm-app-go-template/g'); done
	for f in conf/*.bak; do mv "$$f" $$(echo "$$f" | sed 's/.bak//g'); done

rename-revert-docs:
	for f in docs/raml/*.bak; do mv "$$f" $$(echo "$$f" | sed 's/.bak//g'); done

rename-revert: rename-revert-bamboo-specs rename-revert-conf-files rename-revert-docs
	mv README.md README.template.md
	for f in *.bak; do mv "$$f" $$(echo "$$f" | sed 's/.bak//g'); done
	mv TEMPLATE.md README.md

rename-clean-bamboo-specs:
	if [ -e "bamboo-specs/src/main/java/fusemail/PlanSpec.java.bak" ]; then \
		rm bamboo-specs/src/main/java/fusemail/PlanSpec.java.bak; \
	fi
	if [ -e "bamboo-specs/bamboo.yml.bak" ]; then \
		rm bamboo-specs/bamboo.yml.bak; \
	fi

rename-clean-conf-files:
	find conf/ -name '*.bak' -delete

rename-clean-docs:
	find docs/raml/ -name '*.bak' -delete

rename-clean: rename-clean-bamboo-specs rename-clean-conf-files rename-clean-docs
	find . -name '*.bak' -delete

#########################################
# Help
#########################################
help:
	@echo " $ make	-- creates the debian package (requires local Go env)."
	@echo " $ make build	-- builds binary without creating the debian package (requires local Go env)."
	@echo " $ make test	-- runs go test for all packages (requires local Go env)."
	@echo " $ make test-v	-- runs go test for all packages, verbose (requires local Go env)."
	@echo " $ make docker	-- build final docker run image."
	@echo " $ make docker-build	-- build docker image."
	@echo " $ make docker-artifacts	-- use docker container to build artifacts such as binary file and documentations."
	@echo " $ make docker-push	-- tag and push latest docker image to registry."
	@echo " $ make docker-test	-- run test use docker container."
	@echo " $ make docker-clean	-- remove docker images."
	@echo " $ make docker-run	-- test run docker container (named 'test_run') as deamon service in dev mode with default value set in $(PROJECT_NAME)-dev.env file and exposing port 8080."
	@echo " $ make docker-run-logs	-- tail test run docker container logs."
	@echo " $ make docker-run-help	-- show service usage."
	@echo " $ make docker-run-stop	-- force stop and remove the test run container 'test_run' created by the 'docker-run' target."
	@echo " $ make bamboo-test	-- test bamboo-specs."
	@echo " $ make bamboo-publish	-- publish bamboo-specs."
	@echo " $ make bamboo-clean	-- remove bamboo build target directory."
	@echo " $ make jobspec-tmpl	-- generates per region jobspec template in conf/ directory based on conf/jobspec-devel.hcl file."
	@echo " $ make jobspec	-- creates final jobspec in build/ directory."
	@echo " $ make deploy	-- creates the docker image and deploys it to docker-registry.electric.net and pushes Nomad job specs to artifacts server."
	@echo " $ make clean	-- removes all project related build/test files."
	@echo " $ make rename	-- renames fm-app-go-template into $(PROJECT_NAME) in file names and file contents with original backsup into *.bak files."
	@echo " $ make rename-revert	-- reverts the 'make rename' actions."
	@echo " $ make rename-clean	-- removes *.bak files generated by 'make rename' after confirming all changes are final."
