package server

import (
	"fmt"
	"os"
	"path/filepath"

	"bitbucket.org/fusemail/fm-lib-commons-golang/sys"
	consulapi "github.com/hashicorp/consul/api"
)

// Service ...
type Service struct {
	Name             string
	Tag              string
	ConsulTags       []string
	RegistrationHost string
	Port             int
	UseSSL           bool
	consul           *consulapi.Client
}

func (s *Service) connect() error {
	config := consulapi.DefaultConfig()
	config.Address = s.RegistrationHost
	api, err := consulapi.NewClient(config)
	if err != nil {
		return err
	}
	s.consul = api
	return nil
}

// ConsulName applies default program name if given option name is empty
func ConsulName(consulName string) string {
	if consulName == "" {
		consulName = filepath.Base(os.Args[0])
	}

	return consulName
}

// MustRegister calls Register and on error exits
func (s *Service) MustRegister() {
	if s.consul == nil {
		err := s.connect()
		if err != nil {
			log.WithField("error", err).Error("failed to connect to consul")
			sys.Exit(1)
		}
	}
	err := s.register()
	if err != nil {
		log.WithField("error", err).Error("failed to register service with consul")
		sys.Exit(1)
	}
}

// Register connects to consul and sets up the service and health check
func (s *Service) register() error {
	if s.consul == nil {
		err := s.connect()
		if err != nil {
			return err
		}
	}
	logger := log.WithField("service", "register")

	proto := "http"
	if s.UseSSL {
		proto = "https"
	}
	check := &consulapi.AgentServiceCheck{
		HTTP:     fmt.Sprintf("%s://%s:%d/health", proto, "localhost", s.Port),
		Interval: "5s",
		// MAIL-1090 Per OPS request, remove auto-deregiser
		//	DeregisterCriticalServiceAfter: "15m",
	}

	// MAIL-1209 - custom tags to consul registration
	var tags []string
	if len(s.ConsulTags) == 0 {
		tags = append(tags, "prometheus_exporter")
	} else {
		tags = append(tags, s.ConsulTags...)
	}

	registration := &consulapi.AgentServiceRegistration{
		ID:    s.getID(),
		Name:  s.Name,
		Port:  s.Port,
		Tags:  tags,
		Check: check,
	}

	err := s.consul.Agent().ServiceRegister(registration)

	if err != nil {
		logger.WithField("error", err).Error("failed to register service")
		return err
	}

	logger.Info("registering service")

	return nil
}

// Check implements the health check interface
func (s *Service) Check() (map[string]interface{}, error) {
	return map[string]interface{}{
		"consul_host": s.RegistrationHost,
		"name":        s.Name,
	}, nil
}

func (s *Service) getID() string {
	if s.Tag != "" {
		return fmt.Sprintf("%s-%s", s.Name, s.Tag)
	}
	return s.Name
}

// Deregister ...
func (s *Service) Deregister() error {
	if s.consul == nil {
		return fmt.Errorf("consul api not connected")
	}
	logger := log.WithField("service", "deregister")
	err := s.consul.Agent().ServiceDeregister(s.getID())
	if err != nil {
		logger.WithField("error", err).Error("failed to deregister service")
		return err
	}
	logger.Info("deregistering service")
	return nil
}

// Datacenter get the datacenter name
func (s *Service) Datacenter() (string, error) {
	if s.consul == nil {
		err := s.connect()
		if err != nil {
			return "", fmt.Errorf("failed to connect to consul")
		}
	}

	options := &consulapi.QueryOptions{}
	pair, _, err := s.consul.KV().Get("constants/datacenter/longname", options)
	if err != nil {
		return "", fmt.Errorf("failed to get datacenter: %v", err)
	}

	return string(pair.Value), nil
}
