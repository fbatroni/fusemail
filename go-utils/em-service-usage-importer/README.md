# em-service-usage-importer

This repo establishes a common application/service template for Go-lang development, which includes the following features:

* Basic code framework in ```main.go``` to setup a HTTP service with JSON logging, default endpoints (such as document root ```/```, ```/metrics```, ```/health```, ```/api```, ```/debug/pprof``` etc.), and Consul auto registration.
* Recommended test structure in ```main_test.go``` where example is provided using Go Convey and table drive tests.
* Go package management config files created by [dep](https://golang.github.io/dep/docs/introduction.html) which is currently recommended by Go community.
* ```README.template.md``` file that sets up a service documentation template which will be converted into service's home page at document root ```/``` endpoint. Must be renamed to ```README.md``` file when applying to an actual custom application/service.
* [RAML](https://raml.org) examples under ```docs/raml/``` directory which helps you to get started on documenting API endpoints with some examples under ```docs/raml/examples``` directory.
* Default environment files (```*.env``` files under ```conf/``` directory) which provides a configuration template for this application/service. There should always be a pair of ```.env``` files, one for production environment (```*-prod.env```) and the other for development environment (```*-dev.env```).
* Default Debian package setup files for a service (such as ```conf/em-service-usage-importer.conf```, ```conf/preinstall.sh```, and ```conf/post_install.sh```). These are only needed for traditional deployment without Docker and Nomad.
* Dockerfile that will allow you to setup a build/dev env for Go-lang without having to install Go on your own, and also allows you to build the final deployable image and push to common registry.
* Makefile that helps in building Debian package, binary, documentation, Docker image, test, run, push, and clean.
* Nomad job specs that allows Nomad to deploy this service to specified region and datacenters under certain configuration. (Check out our conventions [here](https://fusemail.atlassian.net/wiki/spaces/SO/pages/223379654/Docker+Nomad+CI+CD+Fabio))
* Bamboo specs that help publish/update build plan configuration to Bamboo for continuous integration and deployment.

*Note:*

* *This README is written based on commands applicable in Linux/Unix/Mac OS system, Windows users please revise accordingly.*
* *All commands listed in this document are assumed to be run from repo root directory unless otherwise specified.*

Related CI/CD conventions are availble on [this Confluence page](https://fusemail.atlassian.net/wiki/spaces/SO/pages/223379654/Docker+Nomad+CI+CD+Fabio).

## Get started

### Dependencies & Preparation

Before start using this template, make sure that:

* [Docker](https://docs.docker.com/install/) is installed.
* [Docker registry](http://docker-registry.electric.net:10080) is accessible through VPN on bastion22. (See more details on [this Confluence page](https://fusemail.atlassian.net/wiki/spaces/SO/pages/223379654/Docker+Nomad+CI+CD+Fabio))
* Your public key (usually at ```~/.ssh/id_rsa.pub```) is uploaded to your Bitbucket account "SSH keys" settings to access repositories without password. See [here](https://confluence.atlassian.com/bitbucket/set-up-an-ssh-key-728138079.html) to learn how to generate an SSH key.
* [Bamboo](http://bamboo2203a.electric.net:8085) is accessible through VPN on bastion22 and you have a username/password to login with. (See service info [here](https://fusemail.atlassian.net/wiki/spaces/SO/pages/108794333/Bamboo))
  * If you are using Bamboo Java Spec, then you need to create a ```bamboo-specs/.credential``` file with your login credentials in the following format:

            # cat bamboo-specs/.credential
            username=admin
            password=admin

### Try out template features

1. Make sure everything listed in "Dependencies & Preparation" section is ready.
1. Clone this repository into your local workstation:

        git clone git@bitbucket.org:fusemail/em-service-usage-importer.git

1. Run test using Docker container:

        make docker-test

1. Build Docker image with associated documentations:

        make docker

1. Check service usage:

        make docker-run-help

1. Run template service using Docker container:

        make docker-run

    This should bring up a web interface at [localhost:8080](http://localhost:8080) where you can access the following endpoints:

    * [Health](/health)
    * [Metrics](/metrics)
    * [System Variables](/sys)
    * [Version](/version)
    * [API Documentation](/api)
    * [Text Log](/log)
    * [JSON Log](/log?format=json)
    * [Debugging](/debug/pprof/)

1. Check container logs:

        make docker-run-logs

1. Stop and remove Docker container:

        make docker-run-stop

1. Push docker image to registry:

        make docker-push

1. (For Java specs only) Test Bamboo Specs:

        make bamboo-test

1. (For Java specs only) Publish Bamboo Specs:

        make bamboo-publish

1. Clean up build/test files:

        make clean

### Make a command-line app using this template

Although the default code framework is written for a HTTP service, you can easily convert it to a command-line tool by removing the HTTP service related features and Consul auto-registration from ```main.go``` file (search for "command-line tool" in comments for detailed instruction).

### Create your own project using this template

1. Create your own repository in Bitbucket, say REPO_NAME, and clone it to your local environment.
1. Run ```make clean``` in this repo and copy all files under this repository (including .gitignore file) into your new repo root directory.
1. Check-in all changes to repo:

        # git add --all
        # git commit -m "initial commit"
        # git push -u origin master

1. If this is a command line utility, then:

        # mv Makefile.cmd Makefile

   If this is a stateful service (i.e. writes to somewhere), then remove ```[ "$$dc" = "devel" ] ||``` condition from the following in "deploy" target:

        if [ "$$dc" = "devel" ] || [ "$(GITBRANCH)" = "master" ];

1. Update the following values in the Makefile:

        // under Package Metadata section
        REPO_NAME=em-service-usage-importer        // set to your own REPO_NAME
        MAINTAINER=developers@fusemail.com  // set to your team's email address
        DESCRIPTION='Service description goes here.'    // set to your own service description

        // under CI/CD Config section
        BUILD_ENV_VERSION=1.7-go1.10.3-stretch   // use latest REGISTRY_ADDR/fusemail/fm-utility-go-build tag
        DEBIAN_VERSION=9.4      // use latest REGISTRY_ADDR/debian/debian tag
        BAMBOO_PROJECT=Templates    // checkout existing projects in [Bamboo](http://bamboo2203a.electric.net:8085/allPlans.action)
        BAMBOO_PROJECT_KEY=TMPLTS   // use the same project key in Babmoo (i.e. the resource name in project URL)
        BAMBOO_PLAN_KEY=FAGT    // use the first letter of each word in REPO_NAME to make your own plan key

        // if this is a service,
        DATACENTERS=fusepoint ireland london sweden1 sweden2 toronto denmark1 // change to your own list of datacenters for deployment if necessary

1. Follow instruction in [How to add webhook to repository] section in this README to enable Bamboo spec auto scanning and Bamboo plan auto update for this repo.

1. By default, this template applies Bamboo Yaml specs so you could remove the Java spec directory below:

        # rm -rf bamboo-java-specs

    (Optional) If you need more features that the default bamboo.yml file does not provide (such as creating project, script tasks dependencies, stage and job naming), then you should use Bamboo Java specs by doing the following renaming:

        # mv bamboo-specs bamboo-yaml-specs
        # mv bamboo-java-specs bamboo-specs

      * Publish Bamboo plan for the first time (if you already setup webhook above then you don't have to do this):

            # make bamboo-publish

        Check out the new plan created on [Bamboo Build Dashboard](http://bamboo2203a.electric.net:8085/allPlans.action).

1. If this repo is a service, then run the following command to generate a jobspec template for each datacenter that you specified above:

        # make jobspec-tmpl

        *Note: If the settings in jobspec need to be changed, you can make change in jobspec-devel.hcl file and run the above command again to populate the jobspec context into other datacentre files.*

1. Run the following commands to rename em-service-usage-importer to your REPO_NAME:

        # make rename           // rename target renames file, and change files inline which also backs up .bak files
        # make rename-clean     // check the .bak files before running this
        # make rename-revert    // (optional) run this to revert the change if anything goes wrong

1. Commit changes in repo again:

        # git add --all
        # git commit -m "setup complete"
        # git push origin master

1. Watch Bamboo plan rebuild in Bamboo plan home page:

        http://bamboo2203a.electric.net:8085/browse/{project key}_{plan key}

### Apply this template to an existing project

1. Checkout both em-service-usage-importer repo and your own repo and switch to the branch where you want to apply this template on.
1. Integrate changes in ```Dockerfile``` from this repo, or just copy over if you don't already have one.
1. (If you choose Yaml spec) Copy over ```bamboo-specs``` directory to your repo and update the following fields in ```bamboo-specs/bamboo.yml``` file accordingly:

        project:
        key: {Your Project Key}
        plan:
            key: {Your Plan Key}
            name: {Your Repo Name}

1. (If you choose Java spec) Copy over ```bamboo-java-specs``` directory to your repo, rename it to ```bamboo-specs```, and update the following fields in ```bamboo-specs/src/main/java/fusemail/PlanSpec.java``` file accordingly:

        public static final String PROJECT_NAME = "YourProjectName";
        public static final String PROJECT_KEY = "YourProjectKey";
        public static final String REPO_NAME = "your-repo-name";
        public static final String PLAN_KEY = "YourPlanKey";

    Make sure you update the bamboo-specs/.credentials file to include your own username/password to access Bamboo.

1. Add this repository as "Linked Repository" in Bamboo and create a Bitbucket webhook according to the instruction in [How to add webhook to repository] section in this README.
1. Integrate changes in ```Makefile``` from this repo (except the ```rename``` targets as they are not needed in this case).
    * If this is a command line utility, then integrate changes from ```Makefile.cmd``` file.
    * If this is a service app, then make sure you change the default port number in ```docker-run``` target to match your own service port number.
    * If this is a stateful service (i.e. writes to somewhere), then remove ```[ "$$dc" = "devel" ] ||``` condition from the following in "deploy" target:

            if [ "$$dc" = "devel" ] || [ "$(GITBRANCH)" = "master" ];

1. Copy over ```conf/*.env``` files from this repo and rename it to ```conf/{REPO_NAME}-{prod|dev}.env``` in your own repo if you don't already have them.
1. (To build Debian package) Copy over ```conf/*.sh``` files from this repo to your own if you don't already have them then update accordingly.
1. If this is a service app, copy over ```conf/jobspec-devel.hcl``` file from this repo into your own and update them accordingly (make sure you change the port number if it's a service). Make copies for other regions deployment, using the following command:

        # make jobspec-tmpl

    Our existing regions are: ```devel, fusepoint, ireland, london, sweden1, sweden2, toronto, denmark1```. Configure your own deployment list in ```Makefile:jobspec-tmpl``` target. Make sure you allocate the suitable amount of CPU and memory for your application/service in the jobspec, do NOT just use the default ones from this template.

1. If your repo is not using Go dep, then convert to [dep](https://github.com/golang/dep). (See instructions [here](https://golang.github.io/dep/docs/migrating.html))
1. Make sure all ```[[constraint]]``` in ```Gopkg.toml``` file which refers to one of our own repository under ```bitbucket.org/fusemail``` has a source setting like the example below so that ```dep``` commands will work within Docker build env:

          source = "git@bitbucket.org:fusemail/fm-lib-commons-golang.git"

1. Make sure vendor directory in your repo is checked in.
1. Integrate changes in ```README.template.md``` file from this repo into your own repo's ```README.md``` file.
1. Copy over ```main_test.go``` file from this repo into your own repo and create tests using the provided framework if you don't currently have a test yet.
1. Run through all targets in ```Makefile``` to make sure the integration works.
    * (If you use Bamboo Java Spec without webhook setup) Especially make sure you run "make bamboo-publish" if Bamboo plan for your project has not been created.

## How to develop using Docker container

[Docker](https://www.docker.com) enables true independence between applications and infrastructure, so that you can now develop/test/run a Go program without having to install any Go dependency on your own.

### Access Docker container

To start a Docker container with Go-lang's build environment and mount your repo directory into the container, simply run the following command:

    # docker run --rm -it -v ~/.ssh:/root/.ssh -v $PWD:/go/src/bitbucket.org/fusemail/em-service-usage-importer docker-registry.electric.net:10080/fusemail/fm-utility-go-build:debian-latest /bin/sh

*Note: If this is your first time using the fm-utility-go-build:debian-latest image, then this command will take a while to download it from docker-registry.electric.net first.*

This should bring you into the container's shell interface, where you can switch to your repo root directory and start working from there:

    /go # cd src/bitbucket.org/fusemail/em-service-usage-importer/

Inside the container, it's as if you have already installed Go-lang and you can run all Go tool and commands. And any changes that you made to the mounted repo root directory will reflect back to your local file system where you can manage with version control system.

To exit the container, just type:

    /go/src/bitbucket.org/fusemail/em-service-usage-importer # exit

Since we used "--rm" option in "docker run" command, this temporary container will be deleted automatically once you have exited the container.

### Manage vendor directory

Use [dep](https://golang.github.io/dep/docs/daily-dep.html) to manage your vendor directory which should then be checked-in to your repository as required by our company convention.

Since *dep* needs to access Bitbucket repository where your ssh key is required, if you have generated your SSH key with a passphrase, then you will need to do the following before running any *dep* commands: (cited from [here](https://superuser.com/questions/988185/how-to-avoid-being-asked-enter-passphrase-for-key-when-im-doing-ssh-operatio))

    /go/src/em-service-usage-importer # eval `ssh-agent -s`; ssh-add ~/.ssh/id_rsa  // enter your passphrase once
    /go/src/em-service-usage-importer # ssh-add -l             // confirm that your key has been added

Use dep init to initiate a manifest:

    /go/src/em-service-usage-importer # dep init

This command should generate Gopkg.toml, Gopkg.lock files, and vendor/ directory which is already done for you in this template.

To check the vendor status, use:

    /go/src/em-service-usage-importer # dep status

To update vendor directory, check out the following command options:

    /go/src/em-service-usage-importer # dep ensure --help

It is required to check in the entire vendor directory according to company convention so that it is easier to set up continuous integration/deployment process as the next step.

### Run test

Use the following command to test all available test files under sub-directories as well:

    /go/src/em-service-usage-importer # go test -v ./...

## How to use Makefile

Makefile provides a convinient way to wrap all the shell scripts that are used to build, test, run and clean this application into one ```make``` program. Check out the "help" section in the bottom of the ```Makefile``` file for usage details, or run ```make help``` command.

## How to use Nomad jobspecs for deployment

The [Nomad](https://www.nomadproject.io) [job specification](https://www.nomadproject.io/docs/job-specification/index.html) (or "jobspec" for short) defines the schema for Nomad jobs. They are specified in [HCL](https://github.com/hashicorp/hcl) which is a simple configuration language developed by HashiCorp.

We store all the jobspec template files for each region in ```conf/``` directory, which can be updated to contain the exact build tags that passed test in Bamboo. The updated jobspecs will be placed under ```build/``` directory and uploaded to artifacts server. This process is achieved by the ```make deploy``` command which can only be executed in Bamboo server for access to the artifact server.

Once these jobspecs are uploaded to the artifacts server, the deploymenator should pick it up and auto deploys to dev environment. This works for all branch artifacts as long as master artifacts. But only master artifacts will enventually make it to the production environment. Currently production deployment doesn't happen automatically unless particularly specified in deploymenator's config, otherwise it will requires an admin person to go and manually call the deployment action to go to production.

In case of deployment failure, you can find out your deployment status from the deploymenator's queue endpoint, e.g. http://10.99.1.146:10050/deployment?state=all (this endpoint can be looked up through [Nomad UI in dev env](https://core9a.electric.net:4646/ui/jobs/fm-service-deploymenator/fm-service-deploymenator-group)). And then debug it on a Nomad server (e.g. core9a/b/c in dev env) host by executing the following commands in order to deploy the application/service and check deployment status:

    // all job specs should be in the following directory
    core9a ~ # mkdir -p /etc/nomad-jobs/em-service-usage-importer
    core9a ~ # cd /etc/nomad-jobs/em-service-usage-importer/

    // download the latest job spec from artifacts server (UI available [here](https://bamboo2203a.electric.net:8098))
    core9a em-service-usage-importer # rm VERSION; wget https://artifacts.electric.net:8098/em-service-usage-importer/latest/VERSION
    core9a em-service-usage-importer # rm jobspec-devel.hcl; wget https://artifacts.electric.net:8098/em-service-usage-importer/$(cat VERSION)/jobspec-devel.hcl

    // prepare to run nomad commands
    core9a em-service-usage-importer # alias nomad='envdir /etc/nomad-management/env nomad'

    // deploy by nomad using jobspec
    core9a em-service-usage-importer # nomad job run jobspec-devel.hcl

    // useful commands to debug deployment
    core9a em-service-usage-importer # nomad job status em-service-usage-importer    // check job's deployment status and allocations
    core9a em-service-usage-importer # nomad node-status <Node ID>    // shows actual host name of the node and jobs that are running on it
    core9a em-service-usage-importer # nomad alloc-status <Allocation ID>     // displays status and metadata about an existing allocation and its tasks.
    core9a em-service-usage-importer # nomad status <Any ID>          // convenient command to show status and metadata of any Node/Allocation ID

    // consule registration check
    core9a em-service-usage-importer # curl em-service-usage-importer.service.consul:8080/health

    // view logs
    nomad9a/b/c ~ # tail -f /var/log/remote-job/em-service-usage-importer.local2.log // login to the actual host provided by the above commands
    syslog9a ~ # ls /var/syslog/remote-job/em-service-usage-importer.local2/em-service-usage-importer.local2.<yyyy-mm-dd>.log  // go to centralized syslog server

    // to update a job
    core9a em-service-usage-importer # nomad job plan jobspec-devel.hcl   // get the Job Modify Index number from this command first
    core9a em-service-usage-importer # nomad job run -check-index <Job Modify Index> jobspec-devel.hcl

    // to stop a job
    core9a em-service-usage-importer # nomad job stop em-service-usage-importer  // this should stop the service and deregister from Consul

### Useful links

* [Nomad UI](https://core9a.electric.net:4646/ui/) in dev env.
* [Consul UI](http://core9a.electric.net:8500/ui/#/devel/nodes) in dev env.
* [Docker Registry UI](https://bamboo2203a.electric.net:10081)
* [Artifacts Server UI](https://bamboo2203a.electric.net:8098)
* See more Nomad commands [here](https://www.nomadproject.io/docs/commands/index.html). Starting from [the tutorial](https://www.nomadproject.io/intro/getting-started/jobs.html) is recommended.

### How to add Vault secrets to jobspec

Sometimes you do need your job to access DB or something that relies on a secret. In this case, we use Vault to store the secrets and have Nomad jobspec pulling these secrets from Vault and set them as env variables.

To see a working example jobspec on how Vault secrets are loaded, check out [this jobspec file](https://bitbucket.org/fusemail/fm-app-verification/src/master/conf/jobspec-devel.hcl). E.g. add the following section to your task section in jobspec.hcl file:

        vault {
          policies = ["###REPO_NAME###"]
          change_mode   = "restart"
        }

        template {
          data = <<EOH
        {{ with secret "secret/mysql/aka/###REPO_NAME###" -}}
        DB_USER="{{ .Data.user }}"
        DB_PASS="{{ .Data.password }}"
        {{ end -}}
        EOH

          destination = "secrets/###REPO_NAME###.env"
          env = true
          perms = "640"
        }

        env {
          "DB_ADDR" = "aka9a.electric.net"
        }

To see how to create Vault secrets in dev env, check out [this ticket comment](https://fusemail.atlassian.net/browse/AP-3943?focusedCommentId=112266&page=com.atlassian.jira.plugin.system.issuetabpanels%3Acomment-tabpanel#comment-112266). You need to set up both vault secrets and vault policy in dev env first, make sure you check-in your vault-policy.hcl file into your repo/conf directory. E.g.

        # To set up vault secret
        vault9a ~ # export VAULT_CACERT=/etc/vault.d/ssl/vault.pem; export VAULT_ADDR=https://vault9a.electric.net:8200/
        vault9a ~ # vault status        // make sure it's not sealed
        vault9a ~ # cd /home/chunyang.chen
        vault9a chunyang.chen # cat DB_HOST | tr -d '\n' > DB_HOST2; mv DB_HOST2 DB_HOST
        vault9a chunyang.chen # cat DB_USER | tr -d '\n' > DB_USER 2; mv DB_USER 2 DB_USER
        vault9a chunyang.chen # cat DB_PASS | tr -d '\n' > DB_PASS2; mv DB_PASS2 DB_PASS
        vault9a chunyang.chen # vault write secret/mysql/aka/###REPO_NAME### host=@DB_HOST user=@DB_USER password=@DB_PASS
        vault9a chunyang.chen # vault read secret/mysql/aka/###REPO_NAME###

        # To set up vault policy
        vault9a ~ # cd /home/chunyang.chen/
        vault9a chunyang.chen # vim vault-policy.hcl
        vault9a chunyang.chen # vault policy write ###REPO_NAME### vault-policy.hcl
        vault9a chunyang.chen # vault policy read ###REPO_NAME###
        path "secret/mysql/aka/###REPO_NAME###" {
                capabilities = ["read"]
        }

See our convention about Vault usage in [this documentation](https://fusemail.atlassian.net/wiki/spaces/SO/pages/381780105/Vault).

## How to use Bamboo for Continous Integration & Deployment

[Atlassian Bamboo](https://www.atlassian.com/software/bamboo) is a CI/CD server used by our company. See [this Confluence page](https://fusemail.atlassian.net/wiki/spaces/FDT/pages/109318543/How+To+Bamboo) for access info.

Bamboo provides a [Bamboo Specs Feature](https://confluence.atlassian.com/bamboo/bamboo-specs-894743906.html) which allows you to store Bamboo build plan's configuration as code in the repository. Make sure you choose the right spec to use for your own case:

| Feature | Java Spec | Yaml Spec |
| ------- |:---------:| ---------:|
| Create Project | Yes | No |
| Create/Update Plan | Yes | Yes |
| Create sequential tasks under a job | Yes | No - only one script task is allowed per job |
| Create final tasks under a job | Yes | No |
| Set stage/job name and description | Yes | No |
| Bitbucket web hook to allow repo push by repo ID | Yes | Yes |
| Branch management | Yes | Yes |
| Auto create plan from specified branch in Linked Repository | Yes | Yes |
| Set notification | Yes | No - default to committer and plan watcher |
| Custom Trigger | Yes | No - default to bitbucket webhook trigger |
| Set plan permission | Yes | No - inherits project permission |
| Generate artifacts | Yes | Yes |

With the Bamboo specs included in your repository, you could [setup a webhook in your Bitbucket repo](http://docs.atlassian.com/bamboo/docs-0606/Enabling+webhooks) which will then push the repo update event to Bamboo and trigger the plan or plan branch to be created/updated and run.

### How to add webhook to repository

Adding webhook from Bitbucket Cloud to Bamboo will enable plan auto creation/update upon repository commits. **You must be the owner or admin of the Bitbucket repo in order to do this.**

Follow the instruction below to add a webhook to your Bitbucket repository:

* Go to [Bamboo UI](http://bamboo2203a.electric.net:8085/admin/configureLinkedRepositories!doDefault.action) to create your repo as a Linked Repository. Then:
* Edit the linked repository, select "Bamboo Specs" tab and make sure "Scan for Bamboo Specs" switch is turned on, and enable "Access to all projects".
* Follow the instruction under "[Setup webhook](http://docs.atlassian.com/bamboo/docs-0606/Enabling+webhooks)" section to set up webhook in Bitbucket for your repo.
  * *Note that you must change the URL domain to [https://bamboo.electric.net](https://bamboo.electric.net) for the setting to work in Bitbucket.*

An example webhook with title "Bamboo build trigger" has been added for em-service-usage-importer, which can be referred to at [this link](https://bitbucket.org/fusemail/em-service-usage-importer/admin/addon/admin/bitbucket-webhooks/bb-webhooks-repo-admin)

Once the Bamboo plan is created for the project and with the Webhook correctly setup, every subsequent commit push from this repository will trigger a plan update and rebuild.

### Create Bamboo specs template

#### Create a Yaml spec

Follow the reference [here](https://confluence.atlassian.com/bamboo/bamboo-yaml-specs-938844479.html) to learn how to create a proper Yaml spec.

#### Create a Java spec

The following command will generate the required bamboo-specs directory in ```bamboo-specs``` directory (currently renamed as ```bamboo-java-specs``` directory):

    # docker run --rm -v $PWD:/java/src/bamboo-project -it docker-registry.electric.net:10080/fusemail/fm-utility-java-build:latest /bin/sh
    ---- following commands are run inside container shell env ----
    / # cd /java/src/bamboo-project
    /java/src/bamboo-project # mvn archetype:generate -B \
        -DarchetypeGroupId=com.atlassian.bamboo -DarchetypeArtifactId=bamboo-specs-archetype \
        -DarchetypeVersion=6.6.1 \
        -DgroupId=com.atlassian.bamboo -DartifactId=bamboo-specs -Dversion=1.0.0-SNAPSHOT \
        -Dpackage=fusemail -Dtemplate=minimal

*Note: make sure "-DarchetypeVersion=6.6.1" option value matches current Bamboo instance version otherwise build will fail.*

The files generated in the ```bamboo-specs/``` directory can be compiled using the container created by the same ```docker-registry.electric.net:10080/fusemail/fm-utility-java-build:latest``` image and publish to our Bamboo instance to configure plans, jobs and tasks.

This step is already done in this template repo, so that we can include the template PlanSpec.java file in the following location:

    bamboo-specs/src/main/java/fusemail/PlanSpec.java

If you need to change the version number, then please edit the ```bamboo-specs/pom.xml``` file.

##### Edit Bamboo Java specs

To edit the specs, import the bamboo-specs directory as a project to a Java IDE (such as Eclipse or IntelliJ), which will take a while to download all the dependencies to setup the environment initially.

Check out Bamboo specs documentation [here](https://docs.atlassian.com/bamboo-specs-docs/6.0.0/) to learn about all the features available.

This template example is created by following this [tutorial](https://confluence.atlassian.com/bamboo/tutorial-create-a-simple-plan-with-bamboo-java-specs-894743911.html).

[Best practice](https://confluence.atlassian.com/bamboo/best-practices-894743915.html) have been applied in the example code.

*Note that if you use this generate/update Bamboo plans, all future changes should happen from the repository first. If any custom changes were made through Bamboo then you should run ```make bamboo-publish``` again to resume the auto updates.*

##### Export existing plan Java specs from Bamboo

In case you already have a plan setup in Bamboo, you can export it from "Plan Configuration" screen under "Actions > View Plan as Bamboo Specs" using bamboo admin account (NOT your LDAP account).

The exported Java code can then be integrated to the existing PlanSpec.java file. Make sure you follow the to-do notes in the exported code as not all information can be exported.

Bamboo only supports Java specs export.

##### Test plan

Only Java specs can be tested with the following command:

    # make bamboo-test

##### Publish plan

At creation of this project, run the following command to create the corresponding Bamboo plan if webhook isn't setup:

    # make bamboo-publish

If the plan already exists, then this command will update the plan. Otherwise, it will create the plan.

If you are using a newly created branch in this repository, this action will create a plan branch accordingly. If [a webhook for this repo in Bitbucket](https://confluence.atlassian.com/bamkb/how-to-trigger-a-bamboo-build-from-bitbucket-cloud-using-a-webhook-872271665.html) is set up, then commiting the new branch will trigger the plan branch build automatically without having to manually publish the specs.

### How to manage plan branches

With [plan branches](https://confluence.atlassian.com/bamboo/using-plan-branches-289276872.html) in Bamboo:

* Any new branch created in the repository can be automatically built and tested using the same build configuration as that of the parent plan.
  * If you modified plan configuration in the branch, then you must create a new plan in order to test the plan update.
* Any branches deleted from the repository can be deleted automatically from Bamboo (by daily cleanup process) according to the settings.
* (*Only available by specifying such permission in Java spec*) You have the flexibility to individually configure branch plans, by overriding the parent plan, if required.
* (*Only available by specifying such permission in Java spec*) Optionally, changes from the feature branch can be automatically merged back to the "master" (e.g. trunk, default or mainline branch) when the build succeeds.