# Nomad jobs are specified in HCL(https://github.com/hashicorp/hcl)
# jobspec reference: https://www.nomadproject.io/docs/job-specification/index.html
# Check out more examples including how to use Vault secrets
# from this repo: https://bitbucket.org/fusemail/fm-service-nomad-sample
job "###NOMAD_NAME###" {

  # Specify this job should run in the region named "devel". 
  # Regions are defined by the Nomad servers' configuration.
  region = "devel"
  # Spread the tasks in this job amongst given array of datacenters.
  datacenters = ["devel"]
        
  # Run this job as a "service" type. Each job type has different
  # properties. See the documentation below for more examples.
  type = "service"
      
  # The update stanza specifies the group's update strategy
  # such as rolling update and canary deployment.
  update {
    # Specify this job to have rolling updates, one-at-a-time
    max_parallel = 1

    # Specifies the minimum time the allocation must be in the healthy state 
    # before it is marked as healthy and unblocks further allocations from being updated. 
    min_healthy_time = "10s"

    # Specifies the deadline in which the allocation must be marked as healthy 
    # after which the allocation is automatically transitioned to unhealthy.
    healthy_deadline = "5m"

    # Specifies if the job should auto-revert to the last stable job on
    # deployment failure. A job is marked as stable if all the allocations
    # as part of its deployment were marked healthy.
    auto_revert = false

    # Specifies that changes to the job that would result in destructive updates
    # should create the specified number of canaries without stopping any previous
    # allocations. Once the operator determines the canaries are healthy, they can
    # be promoted which unblocks a rolling update of the remaining allocations at
    # a rate of max_parallel.
    # This should be default to 1 once our tooling is ready.
    canary = 0
  }

  # A group defines a series of tasks that should be co-located
  # on the same client (host). All tasks within a group will be
  # placed on the same host.
  group "###NOMAD_NAME###-group" {

    # Specify the number of these tasks we want.
    count = 1

    # Amount of space to reserve for /local, /secrets, and /alloc. This is not enforced on disk,
    # but is used for scheduling. Default is 300MB. We set this lower so as not to cause problems
    # when /var/lib/nomad is on a small filesystem. Usually there's not very much in these
    # directories anyhow, just an env file in secrets for a few KB and logs that we don't use
    # since we log to syslog anyhow. Most jobs will not need to change this.
    ephemeral_disk {
        size = "10"
    }

    # Restarts happen on the client that is running the task.
    # This automatically starts a service after it has exited, much like supervisor.
    # The config below will restart up to three times in five seconds,
    # then wait one second before trying again.
    restart {
      attempts = 3
      interval = "5s"
      delay = "1s"
      mode = "delay"
    }
      
    # Create an individual task (unit of work). This particular
    # task utilizes a Docker container to run as a HTTP service.
    task "###NOMAD_NAME###" {
      
      # We only use the docker driver.
      driver = "docker"
      
      # Configuration is specific to each driver.
      config {

        image = "docker-registry.electric.net:10080/fusemail/###REPO_NAME###:###BUILDTAG###"

        # Note that we no longer need --consul command line option to run the service,
        # because the jobspec already registers with Consul by task name.

        # use docker run command argument to be explicit about the run environment
        args = ["-e", "dev"]

        port_map {
          http = 8080
        }

        # We use the syslog driver for docker containers running under nomad.
        # You will notice the use of templating here - see https://www.nomadproject.io/docs/runtime/interpolation.html
        # for more information on what you can template into your job spec.
        # See http://www.rsyslog.com/doc/v8-stable/configuration/properties.html
        # for more information on how syslog uses these config values.
        # All logs are centrally stored on one syslog server per datacenter,
        # under the following convention:
        #   /var/syslog/remote-job/{repo_name}.{syslog-facitlity name}/{repo_name}.{syslog-facitlity name}.{yyyy-mm-dd}.log
        # e.g. in dev env, it will be stored at: 
        # syslog2201a:/var/syslog/remote-job/em-service-usage-importer.local2/em-service-usage-importer.local2.2018-07-12.log
        logging {
          type = "syslog"
          config {
            # The "tag" field appears in unstructured log lines.
            # The contents of "tag" up to the first "/" are used to
            # calculate the path to send logs to on the remote syslog server.
            # For most cases you should just leave it as specified here.
            tag = "${NOMAD_TASK_NAME}/nomad-job/${NOMAD_ALLOC_ID}"

            # The "syslog-facility" field must be either "local1" or "local2".
            # "local1" is used for unstructured logs - e.g. for deploying applications
            # we didn't write ourselves for which no JSON logging is available.
            # "local1" will add some context to the beginning of each log line.
            # "local2" is for structured logs. This does NOT add any extra text and
            # is suitable for JSON logs.
            syslog-facility = "local2"
          }
        } 

      } 

      # This is config for Nomad's built-in logging facilities. Nomad keeps a copy of stdout for each
      # task in a group in /alloc. This config says to keep up to one file of up to 1MB in size. The default
      # is 10 files of up to 10MB each. We do not use these logs since we use the syslog driver in docker
      # instead. These are set low to conserve disk space.
      # DO NOT REMOVE OR CHANGE THIS.
      logs {
        max_files     = 1
        max_file_size = 1
      }                                                                                                                                                       
 
      # Specify the maximum resources required to run the task,
      # include CPU, memory, and bandwidth.
      resources {
        # Please note these are intentionally left at the minimum and should be updated by the writer.
        cpu    = 20 # 20 MHz is minimum CPU usage
        memory = 10 # 10MB is minimum MemoryMB value

        network {
          mbits = 10

          # This requests a static port on 8080 on the host. This
          # will restrict this task to running once per host, since
          # there is only one port 8080 on each host.
          port "http" {
            static = 8080
          } 
        } 
      } 

      # The service block tells Nomad how to register this service
      # with Consul for service discovery and monitoring.
      service {
        # Specifies the name this service will be advertised as in Consul.
        # If not supplied, this will default to the name of the job, group,
        # and task concatenated together with a dash, like "docs-example-server".
        # Each service must have a unique name within the cluster. Names must
        # adhere to RFC-1123 §2.1 (https://tools.ietf.org/html/rfc1123#section-2)
        # and are limited to alphanumeric and hyphen characters (i.e. [a-z0-9\-]),
        # and be less than 64 characters in length.
        name = "###NOMAD_NAME###"

        # This tells Consul to monitor the service on the port
        # labelled "http". Since Nomad allocates high dynamic port
        # numbers, we use labels to refer to them.
        port = "http"

        # Specifies the list of tags to associate with this service. DO NOT REMOVE.
        tags = [
          # Enable prometheus metrics, remove this for bash jobs
          "prometheus_exporter",
          # These are use by Prometheus to automatically add labels to metrics 
          "nomad_group_name=${NOMAD_GROUP_NAME}",
          "nomad_task_name=${NOMAD_TASK_NAME}",
          "nomad_alloc_id=${NOMAD_ALLOC_ID}",
          # Use the urlprefix-internal- tag in your Consul service registration configuration
          # to add endpoints to the internal Fabio cluster (to enable api.fusemail.net functionality).
          # TODO: show examples on how it would work.
          # The internal Fabio cluster is used for non-public services,
          # e.g. Grafana, api.fusemail.net. This cluster listens on 443
          # and has an IP address routable from the Fusemail network but not outside of it.
          # Checkout the strip syntax here: http://fabiolb.net/feature/path-stripping/
          "urlprefix-internal-/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}/version strip=/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}",
          "urlprefix-internal-/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}/log strip=/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}",
          "urlprefix-internal-/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}/health strip=/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}",
          "urlprefix-internal-/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}/metrics strip=/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}",
          "urlprefix-internal-/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}/debug strip=/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}",
          "urlprefix-internal-/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}/sys strip=/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}",
          "urlprefix-internal-/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}/api strip=/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_ID}",
        ]

        # Specifies a health check associated with the service.
        # This can be specified multiple times to define multiple checks for the service.
        # At this time, Nomad supports the grpc, http, script1, and tcp checks.
        check {
          address_mode = "host"
          type     = "http"
          path     = "/health"
          interval = "10s"
          timeout  = "2s"

          # This automatically restarts a service when it goes unhealthy
          # (specified in the service block). This is not a correct behavior
          # for many services - usually it's better to write a service that fails fast.
          check_restart {
            limit = 3
            grace = "5s"
            ignore_warnings = false
          }
        }

      } 

    }
  
  }
}
